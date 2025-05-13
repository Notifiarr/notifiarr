#!/bin/bash
# Uninstaller for Notifiarr DSM 7+

set -e

echo "Stopping Notifiarr..."
/usr/local/etc/rc.d/notifiarr.sh stop || true

echo "Removing Notifiarr files..."
rm -f /usr/bin/notifiarr
rm -f /usr/bin/update-notifiarr.sh
rm -f /usr/bin/rotate-notifiarr-logs.sh
rm -rf /etc/notifiarr
rm -f /usr/local/etc/rc.d/notifiarr.sh
rm -f /etc/cron.d/update-notifiarr
rm -f /volume1/data/notifiarr.log*
rm -f /volume1/data/notifiarr-updates.log*

echo "Removing Notifiarr user and sudoers entry..."
sed -i '/notifiarr/d' /etc/sudoers
synouser --del notifiarr || true

echo "Cleanup complete. Notifiarr has been removed from this system."