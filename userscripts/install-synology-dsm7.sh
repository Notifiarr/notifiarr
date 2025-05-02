#!/bin/bash
# Notifiarr DSM 7+ Installer

set -e

echo "Detecting system architecture..."
ARCH=$(uname -m)
case "$ARCH" in
  x86_64 | amd64) 
    ARCH_NAME="x86_64"
    BINARY_NAME="amd64.linux.gz"
    ;;
  aarch64 | arm64) 
    ARCH_NAME="arm64-v8"
    BINARY_NAME="arm64.linux.gz"
    ;;
  armv7* | armv6* | armhf) 
    ARCH_NAME="arm"
    BINARY_NAME="arm.linux.gz"
    ;;
  *) 
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "Fetching latest Notifiarr release metadata..."
RELEASE_JSON=$(curl -sSL https://api.github.com/repos/Notifiarr/notifiarr/releases/latest)
PACKAGE_URL=$(echo "$RELEASE_JSON" | grep -Eo "https://[^\"]+${ARCH_NAME}.*\.tar\.gz" | head -n 1)
PACKAGE=$(basename "$PACKAGE_URL")

if [ -z "$PACKAGE_URL" ]; then
  echo "Failed to find a valid release package for $ARCH_NAME"
  exit 1
fi

echo "Downloading Notifiarr package: $PACKAGE_URL"
cd /tmp
if ! curl -L -o "$PACKAGE" "$PACKAGE_URL"; then
  echo "Failed to download Notifiarr package."
  exit 1
fi

echo "Extracting Notifiarr package..."
if ! tar -xzf "$PACKAGE"; then
  echo "Failed to extract Notifiarr package."
  exit 1
fi
rm -f "$PACKAGE"

echo "Installing or updating Notifiarr binary..."
if [ -f /usr/bin/notifiarr ]; then
  echo "Stopping existing Notifiarr service..."
  pkill -f "/usr/bin/notifiarr" || true
fi
mv usr/bin/notifiarr /usr/bin/notifiarr
chmod +x /usr/bin/notifiarr
[ -d usr ] && rm -rf usr

echo "Creating or updating config and log directories..."
mkdir -p /etc/notifiarr /var/log/notifiarr
CONFIGFILE=/etc/notifiarr/notifiarr.conf
if [ ! -f "${CONFIGFILE}" ]; then
  echo "Generating config file ${CONFIGFILE}"
  echo " " > "${CONFIGFILE}"
  export DN_LOG_FILE="/var/log/notifiarr/app.log"
  export DN_HTTP_LOG="/var/log/notifiarr/http.log"
  /usr/bin/notifiarr --config "${CONFIGFILE}" --write "${CONFIGFILE}.new"
  mv "${CONFIGFILE}.new" "${CONFIGFILE}"
else
  echo "Config file already exists. Skipping generation."
fi

echo "Setting permissions/ownership on: /usr/bin/notifiarr /var/log/notifiarr"
chmod 0755 /usr/bin/notifiarr /var/log/notifiarr
chmod 0750 /var/log/notifiarr
chown -R notifiarr: /var/log/notifiarr /etc/notifiarr

echo "Creating 'notifiarr' user if missing..."
id notifiarr >/dev/null 2>&1 || {
  synouser --add notifiarr "" Notifiarr 0 "" 0
  synouser --disable notifiarr
}
chown -R notifiarr: /var/log/notifiarr /etc/notifiarr

echo "Adding sudoers entry for smartctl..."
sed -i '/notifiarr/d' /etc/sudoers
echo 'notifiarr ALL=(root) NOPASSWD:/bin/smartctl *' >> /etc/sudoers

echo "Creating or updating DSM boot startup script..."
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

echo "Installing or updating daily auto-update cron job..."
CRON_SCRIPT="/usr/bin/update-notifiarr.sh"
CRON_FILE="/etc/cron.d/update-notifiarr"

curl -sSLo "$CRON_SCRIPT" https://raw.githubusercontent.com/Notifiarr/notifiarr/main/userscripts/install-synology.sh
chmod +x "$CRON_SCRIPT"
echo "10 3 * * * root /bin/bash $CRON_SCRIPT >> /dev/null 2>&1" > "$CRON_FILE"

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

echo "Installation or update complete. Notifiarr is running, persistent, and auto-updating."
echo "Please edit /etc/notifiarr/notifiarr.conf and then restart the service with:"
echo "/usr/local/etc/rc.d/notifiarr.sh stop"
echo "/usr/local/etc/rc.d/notifiarr.sh start"