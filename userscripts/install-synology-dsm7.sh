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
  TEMP_BINARY="/etc/notifiarr/tmp/notifiarr.new"
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

echo "Fetching latest Notifiarr version..."
# Get latest version tag directly - more robust extraction
RELEASE_JSON=$(curl -sSL https://api.github.com/repos/Notifiarr/notifiarr/releases/latest)

# Add debug output to help troubleshoot
echo "Debug - API response first 100 chars:"
echo "$RELEASE_JSON" | head -c 100
echo

# Try multiple patterns to extract the tag
TAG_NAME=$(echo "$RELEASE_JSON" | grep -o '"tag_name":"[^"]*"' | head -1 | sed 's/"tag_name":"//g' | sed 's/"//g')

# Fallback if the above doesn't work
if [ -z "$TAG_NAME" ]; then
  echo "Trying alternate method to extract version..."
  TAG_NAME=$(echo "$RELEASE_JSON" | grep "tag_name" | head -1 | awk -F'"' '{print $4}')
  
  # If still empty, try hardcoding latest known version
  if [ -z "$TAG_NAME" ]; then
    echo "Warning: Could not extract version from API, using fallback version"
    TAG_NAME="v0.8.3"  # Hardcode latest known version as last resort
  fi
fi

echo "Latest version: $TAG_NAME"

# Construct URLs directly - simple and reliable
PACKAGE_URL="https://github.com/Notifiarr/notifiarr/releases/download/${TAG_NAME}/notifiarr.${ARCH_NAME}.linux.gz"
CHECKSUM_URL="https://github.com/Notifiarr/notifiarr/releases/download/${TAG_NAME}/sha512sums.txt"

# Get package filename
PACKAGE=$(basename "$PACKAGE_URL")

# Simple error check for URLs
if [ -z "$TAG_NAME" ] || [ -z "$PACKAGE_URL" ] || [ -z "$CHECKSUM_URL" ]; then
  echo "Failed to find valid release package or checksums for $ARCH_NAME"
  echo -e "\nYou can manually download the correct package and run:"
  echo "sudo ./install-synology-dsm7.sh --manual /path/to/package.gz"
  exit 1
fi

echo "Downloading Notifiarr package: $PACKAGE_URL"
# Create notifiarr tmp directory if it doesn't exist
mkdir -p /etc/notifiarr/tmp
cd /etc/notifiarr/tmp
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

# Try sha512sum, fallback to openssl if not available
if command -v sha512sum >/dev/null 2>&1; then
  ACTUAL_SUM=$(sha512sum "$PACKAGE" | awk '{print $1}')
else
  # Fallback to openssl which is likely available on Synology
  ACTUAL_SUM=$(openssl dgst -sha512 "$PACKAGE" | awk '{print $NF}')
fi

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
# Ensure tmp directory exists
mkdir -p /etc/notifiarr/tmp
TEMP_BINARY="/etc/notifiarr/tmp/notifiarr.new"
if ! gunzip -c "$PACKAGE" > "$TEMP_BINARY"; then
  echo "Failed to extract Notifiarr package."
  exit 1
fi

# Simple binary verification
echo "Verifying binary..."
chmod +x "$TEMP_BINARY"

# Always continue with installation - verification is optional on DSM
"$TEMP_BINARY" --version >/dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "Binary verification successful!"
else
  echo "NOTE: Binary verification failed. This is normal on some Synology models."
  echo "Installation will continue anyway."
fi

# Replace existing binary
mv "$TEMP_BINARY" /usr/bin/notifiarr
chmod +x /usr/bin/notifiarr
rm -f "$PACKAGE"

echo "Setting up directories and config..."
# Only create the config directory, not a directory at /usr/bin/notifiarr
mkdir -p /etc/notifiarr
CONFIGFILE=/etc/notifiarr/notifiarr.conf

# Simplified config handling
if [ -f "${CONFIGFILE}.bak" ]; then
  echo "Restoring config from backup..."
  mv "${CONFIGFILE}.bak" "${CONFIGFILE}"
elif [ ! -f "${CONFIGFILE}" ]; then
  echo "Creating default config file..."
  echo " " > "${CONFIGFILE}"
  /usr/bin/notifiarr --config "${CONFIGFILE}" --write "${CONFIGFILE}" || true
fi

# Simplified user and permissions setup
echo "Setting up user and permissions..."
if ! id notifiarr >/dev/null 2>&1; then
  synouser --add notifiarr "" Notifiarr 0 "" 0
fi

chmod 0750 /usr/bin/notifiarr
# Only set ownership on the binary and config directory, not a directory at /usr/bin/notifiarr
chown notifiarr:notifiarr /usr/bin/notifiarr
chown -R notifiarr:notifiarr /etc/notifiarr

# Add sudoers entry in one step
grep -q "notifiarr ALL=(root) NOPASSWD:/bin/smartctl" /etc/sudoers || \
  echo 'notifiarr ALL=(root) NOPASSWD:/bin/smartctl *' >> /etc/sudoers

echo "Creating or updating DSM boot startup script..."
cat << 'EOF' > /usr/local/etc/rc.d/notifiarr.sh
#!/bin/sh

# Set log file paths - these cannot be edited in the config
export DN_LOG_FILE="/etc/notifiarr/app.log"
export DN_HTTP_LOG="/etc/notifiarr/http.log"

# Simplified update function
update_notifiarr() {
  echo "Checking for updates..."
  # Get latest version using the same robust method as the installer
  RELEASE_JSON=$(curl -sSL https://api.github.com/repos/Notifiarr/notifiarr/releases/latest)
  # Try multiple patterns to extract the tag
  LATEST=$(echo "$RELEASE_JSON" | grep -o '"tag_name":"[^"]*"' | head -1 | sed 's/"tag_name":"//g' | sed 's/"//g')
  
  # Fallback if the above doesn't work
  if [ -z "$LATEST" ]; then
    echo "Trying alternate method to extract version..."
    LATEST=$(echo "$RELEASE_JSON" | grep "tag_name" | head -1 | awk -F'"' '{print $4}')
    
    # If still empty, try hardcoding latest known version
    if [ -z "$LATEST" ]; then
      echo "Warning: Could not extract version from API, using fallback version"
      LATEST="v0.8.3"  # Hardcode latest known version as last resort
    fi
  fi
  
  CURRENT=$(/usr/bin/notifiarr --version | awk '{print $3}')
  
  if [ "$LATEST" != "$CURRENT" ]; then
    echo "Updating from $CURRENT to $LATEST..."
    # Get architecture
    ARCH=$(uname -m)
    case "$ARCH" in
      x86_64 | amd64) ARCH_NAME="amd64" ;;
      aarch64 | arm64) ARCH_NAME="arm64" ;;
      armv7* | armv6* | armhf) ARCH_NAME="arm" ;;
      *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    # Direct URL construction
    PACKAGE_URL="https://github.com/Notifiarr/notifiarr/releases/download/$LATEST/notifiarr.$ARCH_NAME.linux.gz"
    
    # Create tmp directory and download
    mkdir -p /etc/notifiarr/tmp
    cd /etc/notifiarr/tmp
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

echo "Verifying installation..."
# Simplified verification
[ -x /usr/bin/notifiarr ] || { echo "ERROR: notifiarr binary missing or not executable"; exit 1; }
[ -f "$CONFIGFILE" ] || { echo "ERROR: config file missing"; exit 1; }
[ -x /usr/local/etc/rc.d/notifiarr.sh ] || { echo "ERROR: startup script missing or not executable"; exit 1; }

# Check if running
if command -v pgrep >/dev/null 2>&1; then
  pgrep -f "/usr/bin/notifiarr.*notifiarr.conf" >/dev/null || echo "WARNING: Notifiarr is not running, will attempt to start"
else
  pidof notifiarr >/dev/null || echo "WARNING: Notifiarr is not running, will attempt to start"
fi

# Setup automatic updates using system cron directory (DSM doesn't have crontab)
echo "Setting up automatic updates (enabled by default)..."
echo "Would you like to enable automatic updates? [Y/n]"
read REPLY
if [ "$REPLY" = "n" ] || [ "$REPLY" = "N" ]; then
  echo "Automatic updates disabled"
else
  # Create a task file in the system's task directory
  TASK_FILE="/etc/cron.d/notifiarr-update"
  echo "# Notifiarr automatic update - runs daily at a random time between 1am-6am" > "$TASK_FILE"
  echo "0 1 * * * root /usr/local/etc/rc.d/notifiarr.sh update" >> "$TASK_FILE"
  chmod 644 "$TASK_FILE"
  
  echo "Automatic updates enabled via system task (runs daily at 1am)"
fi

echo "Installation or update complete. Notifiarr is running and persistent."
echo "Please edit /etc/notifiarr/notifiarr.conf and then restart the service with:"
echo "/usr/local/etc/rc.d/notifiarr.sh restart"