#!/bin/bash
# Notifiarr DSM 7+ Installer

set -e

# Handle command line arguments
MANUAL_PACKAGE=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --manual|-m)
      MANUAL_PACKAGE="$2"
      shift 2
      ;;
    *)
      echo "Unknown argument: $1"
      exit 1
      ;;
  esac
done

if [ -n "$MANUAL_PACKAGE" ]; then
  if [ ! -f "$MANUAL_PACKAGE" ]; then
    echo "Manual package file not found: $MANUAL_PACKAGE"
    exit 1
  fi
  PACKAGE=$(basename "$MANUAL_PACKAGE")
  TEMP_BINARY="/tmp/notifiarr.new"
  echo "Using manually provided package: $MANUAL_PACKAGE"
  if ! gunzip -c "$MANUAL_PACKAGE" > "$TEMP_BINARY"; then
    echo "Failed to extract manual package."
    exit 1
  fi
  # Skip to binary verification
  goto_verify_binary
fi

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
echo "Debug - Raw API response:"
echo "$RELEASE_JSON" | jq . 2>/dev/null || echo "$RELEASE_JSON"

# Get all assets and filter for our architecture
ASSETS=$(echo "$RELEASE_JSON" | jq -r '.assets[] | "\(.name):\(.browser_download_url)"' 2>/dev/null)
echo "Debug - Available assets:"
echo "$ASSETS"

# Find package URL - try multiple patterns
PACKAGE_URL=$(echo "$ASSETS" | grep -iE "notifiarr\.${ARCH_NAME//./\\.}\.linux\.gz" | cut -d: -f2 | head -n 1)
[ -z "$PACKAGE_URL" ] && PACKAGE_URL=$(echo "$ASSETS" | grep -iE "${ARCH_NAME//./\\.}\.linux\.gz" | cut -d: -f2 | head -n 1)

# Find checksums - try both sha512sums.txt and checksums.txt
CHECKSUM_URL=$(echo "$ASSETS" | grep -i "sha512sums\.txt" | cut -d: -f2 | head -n 1)
[ -z "$CHECKSUM_URL" ] && CHECKSUM_URL=$(echo "$ASSETS" | grep -i "checksums\.txt" | cut -d: -f2 | head -n 1)

PACKAGE=$(basename "$PACKAGE_URL")

if [ -z "$PACKAGE_URL" ] || [ -z "$CHECKSUM_URL" ]; then
  echo "Failed to find valid release package or checksums for $ARCH_NAME"
  echo "Available assets:"
  echo "$ASSETS" | sed 's/:/: /'
  echo -e "\nYou can manually download the correct package and run:"
  echo "sudo ./install-synology-dsm7.sh --manual /path/to/package.gz"
  exit 1
fi

echo "Downloading Notifiarr package: $PACKAGE_URL"
cd /tmp
if ! curl -L -o "$PACKAGE" "$PACKAGE_URL"; then
  echo "Failed to download Notifiarr package."
  exit 1
fi

echo "Downloading checksums..."
if ! curl -L -o checksums.txt "$CHECKSUM_URL"; then
  echo "Failed to download checksums file."
  exit 1
fi

echo "Verifying package checksum..."
EXPECTED_SUM=$(grep "$PACKAGE" checksums.txt | awk '{print $1}')
ACTUAL_SUM=$(sha256sum "$PACKAGE" | awk '{print $1}')

if [ "$EXPECTED_SUM" != "$ACTUAL_SUM" ]; then
  echo "Checksum verification failed!"
  echo "Expected: $EXPECTED_SUM"
  echo "Actual:   $ACTUAL_SUM"
  rm -f "$PACKAGE" checksums.txt
  exit 1
fi
rm -f checksums.txt

echo "Checking current installation..."
if [ -f /usr/bin/notifiarr ]; then
  CURRENT_VERSION=$(/usr/bin/notifiarr --version 2>/dev/null | awk '{print $3}' || echo "unknown")
  echo "Existing Notifiarr version detected: ${CURRENT_VERSION:-unknown}"
  
  echo "Stopping existing Notifiarr service..."
  if [ -f /usr/local/etc/rc.d/notifiarr.sh ]; then
    /usr/local/etc/rc.d/notifiarr.sh stop || true
  else
    pkill -f "/usr/bin/notifiarr" || true
  fi
  # Give processes time to terminate
  sleep 10
  
  # Backup existing config if it exists
  if [ -f /etc/notifiarr/notifiarr.conf ]; then
    echo "Backing up existing config..."
    cp -a /etc/notifiarr/notifiarr.conf /etc/notifiarr/notifiarr.conf.bak
  fi
fi

echo "Extracting Notifiarr package..."
TEMP_BINARY="/tmp/notifiarr.new"
if ! gunzip -c "$PACKAGE" > "$TEMP_BINARY"; then
  echo "Failed to extract Notifiarr package."
  exit 1
fi

# Verify new binary works
echo "Verifying new binary..."
chmod +x "$TEMP_BINARY"
if ! "$TEMP_BINARY" --version >/dev/null 2>&1; then
  echo "ERROR: New binary verification failed!"
  rm -f "$TEMP_BINARY"
  exit 1
fi

# Replace existing binary
mv "$TEMP_BINARY" /usr/bin/notifiarr
chmod +x /usr/bin/notifiarr
rm -f "$PACKAGE"

echo "Creating or updating config and log directories..."
mkdir -p /etc/notifiarr /usr/bin/notifiarr
CONFIGFILE=/etc/notifiarr/notifiarr.conf
if [ ! -f "${CONFIGFILE}" ]; then
  echo "Generating config file ${CONFIGFILE}"
  echo " " > "${CONFIGFILE}"
  /usr/bin/notifiarr --config "${CONFIGFILE}" --write "${CONFIGFILE}.new"
  mv "${CONFIGFILE}.new" "${CONFIGFILE}"
elif [ -f "${CONFIGFILE}.bak" ]; then
  echo "Restoring config from backup..."
  mv "${CONFIGFILE}.bak" "${CONFIGFILE}"
else
  echo "Config file already exists. Skipping generation."
fi

echo "Creating 'notifiarr' user if missing..."
if ! id notifiarr >/dev/null 2>&1; then
  synouser --add notifiarr "" Notifiarr 0 "" 0
fi

echo "Setting permissions/ownership on: /usr/bin/notifiarr"
chmod 0750 /usr/bin/notifiarr
chown -R notifiarr:notifiarr /usr/bin/notifiarr /etc/notifiarr

echo "Adding sudoers entry for smartctl..."
if ! grep -q "notifiarr ALL=(root) NOPASSWD:/bin/smartctl" /etc/sudoers; then
  echo 'notifiarr ALL=(root) NOPASSWD:/bin/smartctl *' >> /etc/sudoers
fi

echo "Creating or updating DSM boot startup script..."
cat << 'EOF' > /usr/local/etc/rc.d/notifiarr.sh
#!/bin/sh

# Set log file paths - these cannot be edited in the config
export DN_LOG_FILE="/etc/notifiarr/app.log"
export DN_HTTP_LOG="/etc/notifiarr/http.log"

get_latest_version() {
  curl -sSL https://api.github.com/repos/Notifiarr/notifiarr/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

update_notifiarr() {
  echo "Checking for updates..."
  LATEST=$(get_latest_version)
  CURRENT=$(/usr/bin/notifiarr --version | awk '{print $3}')
  
  if [ "$LATEST" != "$CURRENT" ]; then
    echo "Updating from $CURRENT to $LATEST..."
    ARCH=$(uname -m)
    case "$ARCH" in
      x86_64 | amd64) ARCH_NAME="amd64" ;;
      aarch64 | arm64) ARCH_NAME="arm64" ;;
      armv7* | armv6* | armhf) ARCH_NAME="arm" ;;
      *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    PACKAGE_URL="https://github.com/Notifiarr/notifiarr/releases/download/$LATEST/notifiarr.$ARCH_NAME.linux.gz"
    cd /tmp
    if curl -L -o notifiarr.gz "$PACKAGE_URL"; then
      gunzip -c notifiarr.gz > /usr/bin/notifiarr
      chmod +x /usr/bin/notifiarr
      rm -f notifiarr.gz
      echo "Update complete. Restarting service..."
      $0 restart
    else
      echo "Failed to download update package"
    fi
  else
    echo "Already running latest version: $CURRENT"
  fi
}

case "$1" in
start)
  echo "Starting Notifiarr..."
  /usr/bin/notifiarr -c /etc/notifiarr/notifiarr.conf &
  ;;
stop)
  echo "Stopping Notifiarr..."
  pkill -f "/usr/bin/notifiarr -c /etc/notifiarr/notifiarr.conf"
  ;;
restart)
  echo "Restarting Notifiarr..."
  $0 stop
  sleep 10
  $0 start
  ;;
update)
  # Random delay up to 5 hours (18000 seconds) to spread load
  sleep $((RANDOM % 18000))
  update_notifiarr
  ;;
force-update)
  # Update immediately without random delay
  update_notifiarr
  ;;
esac
exit 0
EOF

chmod +x /usr/local/etc/rc.d/notifiarr.sh
/usr/local/etc/rc.d/notifiarr.sh start

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

read -p "Would you like to enable automatic updates? [Y/n] " -r
if [[ $REPLY =~ ^[Yy]$ ]] || [[ -z $REPLY ]]; then
  echo "Adding daily update cron job..."
  (crontab -l 2>/dev/null; echo "0 1 * * * /usr/local/etc/rc.d/notifiarr.sh update") | crontab -
  echo "OK: Automatic updates enabled (runs daily at 1am)"
else
  echo "Automatic updates disabled"
fi

echo "Installation or update complete. Notifiarr is running and persistent."
echo "Please edit /etc/notifiarr/notifiarr.conf and then restart the service with:"
echo "/usr/local/etc/rc.d/notifiarr.sh restart"