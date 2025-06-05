#!/bin/bash
# Uninstaller for Notifiarr DSM 7+

set -e

if [ "$EUID" -ne 0 ]; then
  echo "This script must be run as root or with sudo."
  exit 1
fi

echo "Stopping Notifiarr service..."
/usr/local/etc/rc.d/notifiarr.sh stop || true

echo "Removing Notifiarr files..."
echo "Checking for /usr/bin/notifiarr:"
[ -f /usr/bin/notifiarr ] && ls -la /usr/bin/notifiarr || echo "File not found"
rm -f /usr/bin/notifiarr

echo "Checking for /usr/bin/update-notifiarr.sh:"
[ -f "/usr/bin/update-notifiarr.sh" ]  || echo "File not found:  /usr/bin/update-notifiarr.sh"

echo "Checking for /etc/notifiarr:"
[ -d /etc/notifiarr ] && ls -la /etc/notifiarr || echo "Directory not found"
if [ -d /etc/notifiarr ]; then
    read -p "Keep Notifiarr config files? [Y/n] " -r
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        echo "Removing all files except notifiarr.conf..."
        find /etc/notifiarr -mindepth 1 ! -name 'notifiarr.conf' -delete
        echo "Preserved notifiarr.conf, removed other files"
    else
        echo "Preserving all config files"
    fi
fi

echo "Checking for /usr/local/etc/rc.d/notifiarr.sh:"
[ -f /usr/local/etc/rc.d/notifiarr.sh ] && ls -la /usr/local/etc/rc.d/notifiarr.sh || echo "File not found"
rm -f /usr/local/etc/rc.d/notifiarr.sh

echo "Checking for /etc/cron.d/update-notifiarr:"
[ -f /etc/cron.d/update-notifiarr ] && ls -la /etc/cron.d/update-notifiarr || echo "File not found"
rm -f /etc/cron.d/update-notifiarr

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
[ ! -d /etc/notifiarr ] && echo "OK: Config directory removed" || echo "WARNING: Config directory still exists"
[ ! -d /volume1/@appdata/notifiarr ] && echo "OK: Log directory removed" || echo "WARNING: Log directory still exists"
! pgrep -f notifiarr >/dev/null 2>&1 && echo "OK: No Notifiarr processes running" || echo "WARNING: Notifiarr process still running"

echo "Cleanup complete. Notifiarr has been removed from this system."