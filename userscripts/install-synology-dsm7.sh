#!/bin/bash
# Notifiarr DSM 7+ Installer

set -e

echo "Detecting system architecture..."
ARCH=$(uname -m)
case "$ARCH" in
  x86_64 | amd64) ARCH_NAME="x86_64" ;;
  aarch64 | arm64) ARCH_NAME="arm64-v8" ;;
  armv7* | armv6* | armhf) ARCH_NAME="armhf" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "Checking for opkg..."
if ! command -v opkg >/dev/null 2>&1; then
  echo "opkg is not installed. Please install Entware: https://github.com/Entware/Entware/wiki/Install-on-Synology-NAS"
  exit 1
fi

echo "Installing zstd via opkg..."
opkg update
opkg install zstd

echo "Verifying zstd..."
if ! command -v zstd >/dev/null 2>&1; then
  echo "zstd installation failed."
  exit 1
fi

echo "Fetching latest Notifiarr release metadata..."
RELEASE_JSON=$(curl -sSL https://api.github.com/repos/Notifiarr/notifiarr/releases/latest)
PACKAGE_URL=$(echo "$RELEASE_JSON" | grep -Eo "https://[^\"]+${ARCH_NAME}.*\.pkg\.tar\.zst" | head -n 1)
PACKAGE=$(basename "$PACKAGE_URL")

if [ -z "$PACKAGE_URL" ]; then
  echo "Failed to find a valid release package for $ARCH_NAME"
  exit 1
fi

echo "Downloading Notifiarr package: $PACKAGE_URL"
cd /tmp
curl -L -o "$PACKAGE" "$PACKAGE_URL"

echo "Extracting Notifiarr package..."
if tar --help | grep -q -- '--zstd'; then
  tar --zstd -xf "$PACKAGE"
else
  FALLBACK_TAR="${PACKAGE%.zst}"
  zstd -d "$PACKAGE" -o "$FALLBACK_TAR"
  tar -xf "$FALLBACK_TAR"
  rm -f "$FALLBACK_TAR"
fi

echo "Installing Notifiarr binary..."
mv usr/bin/notifiarr /usr/bin/notifiarr
chmod +x /usr/bin/notifiarr
rm -rf usr "$PACKAGE"

echo "Creating config and log directories..."
mkdir -p /etc/notifiarr /volume1/data
chown -R root:root /etc/notifiarr /volume1/data
chmod 755 /volume1/data

CONFIGFILE=/etc/notifiarr/notifiarr.conf
echo "Downloading clean default config from GitHub..."
curl -sSLo "$CONFIGFILE" https://raw.githubusercontent.com/Notifiarr/notifiarr/main/examples/notifiarr.conf.example

echo "Fixing log_file path in config..."
sed -i 's|^\(log_file\s*=\s*\).*|\1"/volume1/data/notifiarr.log"|' "$CONFIGFILE"

echo "Creating 'notifiarr' user if missing..."
id notifiarr >/dev/null 2>&1 || {
  PASS=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 32)
  synouser --add notifiarr "$PASS" Notifiarr 0 support@notifiarr.com 0
}
chown notifiarr /volume1/data

echo "Adding sudoers entry for smartctl..."
sed -i '/notifiarr/d' /etc/sudoers
echo 'notifiarr ALL=(root) NOPASSWD:/bin/smartctl *' >> /etc/sudoers

echo "Creating DSM boot startup script..."
cat << 'EOF' > /usr/local/etc/rc.d/notifiarr.sh
#!/bin/sh
case "$1" in
  start)
    echo "Starting Notifiarr..."
    /usr/bin/notifiarr -c /etc/notifiarr/notifiarr.conf &
    ;;
  stop)
    echo "Stopping Notifiarr..."
    pkill -f "/usr/bin/notifiarr -c /etc/notifiarr/notifiarr.conf"
    ;;
esac
exit 0
EOF

chmod +x /usr/local/etc/rc.d/notifiarr.sh
/usr/local/etc/rc.d/notifiarr.sh start

echo "Installing daily auto-update cron job..."
CRON_SCRIPT="/usr/bin/update-notifiarr.sh"
CRON_LOG="/volume1/data/notifiarr-updates.log"
CRON_FILE="/etc/cron.d/update-notifiarr"

curl -sSLo "$CRON_SCRIPT" https://raw.githubusercontent.com/Notifiarr/notifiarr/main/userscripts/install-synology.sh
chmod +x "$CRON_SCRIPT"
echo "10 3 * * * root /bin/bash $CRON_SCRIPT >> $CRON_LOG 2>&1" > "$CRON_FILE"

echo "Installing weekly log rotation script..."
LOGROTATE_SCRIPT="/usr/bin/rotate-notifiarr-logs.sh"
cat << 'EOF' > "$LOGROTATE_SCRIPT"
#!/bin/sh
LOG_DIR="/volume1/data"
LOG_FILE="notifiarr.log"
MAX_BACKUPS=4

cd "$LOG_DIR" || exit 1

if [ -f "$LOG_FILE" ]; then
  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
  cp "$LOG_FILE" "$LOG_FILE.$TIMESTAMP"
  gzip "$LOG_FILE.$TIMESTAMP"
  : > "$LOG_FILE"
fi

# Cleanup old backups
ls -t "$LOG_FILE."*.gz 2>/dev/null | tail -n +$((MAX_BACKUPS + 1)) | xargs -r rm -f
EOF

chmod +x "$LOGROTATE_SCRIPT"
grep -q rotate-notifiarr-logs /etc/cron.d/update-notifiarr || echo "0 4 * * 0 root /bin/sh $LOGROTATE_SCRIPT" >> /etc/cron.d/update-notifiarr

echo "Running final install verification..."
fail() { echo "FAILED: $1"; exit 1; }
check() { [ -x "$1" ] && echo "OK: $1" || fail "$1 missing or not executable"; }
checkfile() { [ -f "$1" ] && echo "OK: $1" || fail "$1 missing"; }

check /usr/bin/notifiarr
if command -v pgrep >/dev/null 2>&1; then
  pgrep -f "/usr/bin/notifiarr.*notifiarr.conf" >/dev/null && echo "OK: Notifiarr is running" || fail "Notifiarr is not running"
else
  pidof notifiarr >/dev/null && echo "OK: Notifiarr is running (via pidof)" || fail "Notifiarr is not running"
fi

checkfile "$CONFIGFILE"
check /usr/local/etc/rc.d/notifiarr.sh
grep -q update-notifiarr "$CRON_FILE" && echo "OK: Cron job is installed" || fail "Cron job not found"
echo "OK: Log file should be at /volume1/data/notifiarr.log"

echo "Installation complete. Notifiarr is running, persistent, auto-updating, and safely logging to /volume1/data."
echo "Please edit /etc/notifiarr/notifiarr.conf and then restart the service with:"
echo "/usr/local/etc/rc.d/notifiarr.sh stop"
echo "/usr/local/etc/rc.d/notifiarr.sh start"