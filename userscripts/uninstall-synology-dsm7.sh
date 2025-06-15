#!/bin/bash
# Uninstaller for Notifiarr DSM 7+

set -e

if [ "$EUID" -ne 0 ]; then
  echo "This script must be run as root or with sudo."
  exit 1
fi

echo "Stopping Notifiarr service and processes..."
# Try multiple methods to ensure all processes are stopped
if [ -f /usr/local/etc/rc.d/notifiarr.sh ]; then
  /usr/local/etc/rc.d/notifiarr.sh stop || true
fi

# Give processes time to terminate
sleep 5

echo "Removing Notifiarr files..."
echo "Checking for /usr/bin/notifiarr:"
if [ -f /usr/bin/notifiarr ]; then
  ls -la /usr/bin/notifiarr
  echo "Removing binary with force..."
  
  # Try to remove with force
  rm -f /usr/bin/notifiarr
  
  # Check if removal was successful
  if [ -f /usr/bin/notifiarr ]; then
    echo "First removal attempt failed, trying with different approach..."
    
    # Try to change permissions first
    chmod 777 /usr/bin/notifiarr 2>/dev/null || true
    
    # Try again with force
    rm -f /usr/bin/notifiarr
    
    # If still exists, inform user
    if [ -f /usr/bin/notifiarr ]; then
      echo "WARNING: Could not remove binary. It might be in use or protected."
      echo "Please manually remove it with: sudo rm -f /usr/bin/notifiarr"
    else
      echo "Successfully removed binary on second attempt."
    fi
  else
    echo "Successfully removed binary."
  fi
else
  echo "Binary not found at /usr/bin/notifiarr"
fi

# Also check for the directory that might have been created
echo "Checking for /usr/bin/notifiarr/ directory:"
if [ -d /usr/bin/notifiarr ]; then
  ls -la /usr/bin/notifiarr/
  echo "Removing directory..."
  rm -rf /usr/bin/notifiarr/
fi

echo "Checking for /usr/bin/update-notifiarr.sh:"
[ -f "/usr/bin/update-notifiarr.sh" ]  || echo "File not found:  /usr/bin/update-notifiarr.sh"

echo "Checking for /etc/notifiarr:"
[ -d /etc/notifiarr ] && ls -la /etc/notifiarr || echo "Directory not found"
if [ -d /etc/notifiarr ]; then
    if [ -f /etc/notifiarr/notifiarr.conf ]; then
        echo "Configuration file found: /etc/notifiarr/notifiarr.conf"
        echo "Would you like to keep the configuration file? [Y/n]"
        read KEEP_CONFIG
        
        if [ "$KEEP_CONFIG" = "n" ] || [ "$KEEP_CONFIG" = "N" ]; then
            echo "Removing all files including configuration..."
            rm -rf /etc/notifiarr
            echo "All configuration files removed"
        else
            echo "Preserving notifiarr.conf and removing all other files..."
            
            # Backup notifiarr.conf
            cp -a /etc/notifiarr/notifiarr.conf /tmp/notifiarr.conf.bak
            
            # Remove everything in the directory
            rm -rf /etc/notifiarr/*
            
            # Restore notifiarr.conf
            mv /tmp/notifiarr.conf.bak /etc/notifiarr/notifiarr.conf
            echo "Preserved notifiarr.conf, removed all other files"
            
            # Make sure tmp directory is gone
            rm -rf /etc/notifiarr/tmp
        fi
    else
        echo "No configuration file found, removing directory..."
        rm -rf /etc/notifiarr
    fi
fi

echo "Checking for /usr/local/etc/rc.d/notifiarr.sh:"
[ -f /usr/local/etc/rc.d/notifiarr.sh ] && ls -la /usr/local/etc/rc.d/notifiarr.sh || echo "File not found"
rm -f /usr/local/etc/rc.d/notifiarr.sh

echo "Checking for cron jobs..."
# Check for the original cron file
echo "Checking for /etc/cron.d/update-notifiarr:"
[ -f /etc/cron.d/update-notifiarr ] && ls -la /etc/cron.d/update-notifiarr || echo "File not found"
rm -f /etc/cron.d/update-notifiarr

# Check for the system cron file
echo "Checking for /etc/cron.d/notifiarr-update:"
[ -f /etc/cron.d/notifiarr-update ] && ls -la /etc/cron.d/notifiarr-update || echo "File not found"
rm -f /etc/cron.d/notifiarr-update

echo "Checking for /volume1/@appdata/notifiarr:"
[ -d /volume1/@appdata/notifiarr ] && ls -la /volume1/@appdata/notifiarr || echo "Directory not found"
rm -rf /volume1/@appdata/notifiarr

echo "Removing Notifiarr user and sudoers entry..."
echo "Removing sudoers entry for smartctl:"
if grep -q "notifiarr ALL=(root) NOPASSWD:/bin/smartctl" /etc/sudoers; then
    echo "Removing sudoers entry for smartctl"
    sed -i '/notifiarr ALL=(root) NOPASSWD:\/bin\/smartctl/d' /etc/sudoers
else
    echo "No smartctl sudoers entry found"
fi

echo "Removing notifiarr user:"
if id notifiarr >/dev/null 2>&1; then
    echo "User details being removed:"
    id notifiarr
    synouser --del notifiarr || true
else
    echo "User not found"
fi

echo "Verifying removal..."
[ ! -f /usr/bin/notifiarr ] && echo "OK: Binary removed" || echo "WARNING: Binary still exists"
[ ! -f /usr/local/etc/rc.d/notifiarr.sh ] && echo "OK: Startup script removed" || echo "WARNING: Startup script still exists"

# Check config directory status
if [ -d /etc/notifiarr ]; then
    if [ -f /etc/notifiarr/notifiarr.conf ]; then
        echo "OK: Configuration file was preserved as requested"
    else
        echo "WARNING: Config directory exists but notifiarr.conf is missing"
        ls -la /etc/notifiarr
    fi
else
    echo "OK: Config directory was completely removed"
fi

[ ! -d /volume1/@appdata/notifiarr ] && echo "OK: Log directory removed" || echo "WARNING: Log directory still exists"
! pgrep -f notifiarr >/dev/null 2>&1 && echo "OK: No Notifiarr processes running" || echo "WARNING: Notifiarr process still running"

echo "Cleanup complete. Notifiarr has been removed from this system."