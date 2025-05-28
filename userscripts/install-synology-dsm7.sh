#!/bin/bash
# Notifiarr DSM 7+ Installer

set -e

# Check for root privileges
if [ "$EUID" -ne 0 ]; then
  echo "This script must be run as root or with sudo."
  exit 1
fi

echo "Detecting system architecture..."
ARCH=$(uname -m)
case "$ARCH" in
  x86_64 | amd64) 
    ARCH_NAME="amd64"
    ;;
  aarch64 | arm64) 
    ARCH_NAME="arm64"
    ;;
  armv7* | armv6* | armhf) 
    ARCH_NAME="arm"
    ;;
  *) 
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "Fetching latest Notifiarr release metadata..."
RELEASE_JSON=$(curl -sSL https://api.github.com/repos/Notifiarr/notifiarr/releases/latest)
PACKAGE_URL=$(echo "$RELEASE_JSON" | grep -o "https://[^\"]*notifiarr\.${ARCH_NAME//./\\.}\.linux\.gz" | head -n 1)
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

echo "Installing or updating Notifiarr binary..."
if [ -f /usr/bin/notifiarr ]; then
  echo "Stopping existing Notifiarr service..."
  pkill -f "/usr/bin/notifiarr" || true
  # Give processes time to terminate
  sleep 2
fi

echo "Extracting Notifiarr package..."
if ! gunzip -c "$PACKAGE" > /usr/bin/notifiarr; then
  echo "Failed to extract Notifiarr package."
  exit 1
fi

chmod +x /usr/bin/notifiarr
rm -f "$PACKAGE"

echo "Creating or updating config and log directories..."
mkdir -p /etc/notifiarr /volume1/@appdata/notifiarr/lognotifiarr
CONFIGFILE=/etc/notifiarr/notifiarr.conf
if [ ! -f "${CONFIGFILE}" ]; then
  echo "Generating config file ${CONFIGFILE}"
  echo " " > "${CONFIGFILE}"
  /usr/bin/notifiarr --config "${CONFIGFILE}" --write "${CONFIGFILE}.new"
  mv "${CONFIGFILE}.new" "${CONFIGFILE}"
else
  echo "Config file already exists. Skipping generation."
fi

echo "Creating 'notifiarr' user if missing..."
if ! id notifiarr >/dev/null 2>&1; then
  synouser --add notifiarr "" Notifiarr 0 "" 0
fi

echo "Setting permissions/ownership on: /usr/bin/notifiarr /volume1/@appdata/notifiarr/lognotifiarr"
chmod 0750 /usr/bin/notifiarr /volume1/@appdata/notifiarr/lognotifiarr
chown -R notifiarr:notifiarr /volume1/@appdata/notifiarr/lognotifiarr /etc/notifiarr

echo "Adding sudoers entry for smartctl..."
if ! grep -q "notifiarr ALL=(root) NOPASSWD:/bin/smartctl" /etc/sudoers; then
  echo 'notifiarr ALL=(root) NOPASSWD:/bin/smartctl *' >> /etc/sudoers
fi

echo "Creating or updating DSM boot startup script..."
cat << 'EOF' > /usr/local/etc/rc.d/notifiarr.sh
#!/bin/sh

# Set log file paths - these cannot be edited in the config
export DN_LOG_FILE="/volume1/@appdata/notifiarr/lognotifiarr/app.log"
export DN_HTTP_LOG="/volume1/@appdata/notifiarr/lognotifiarr/http.log"

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

curl -sSLo "$CRON_SCRIPT" https://raw.githubusercontent.com/Notifiarr/notifiarr/main/userscripts/install-synology-dsm7.sh
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