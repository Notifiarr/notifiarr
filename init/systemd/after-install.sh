#!/bin/sh

# This file is used by deb, rpm and BSD packages.
# FPM/nFPM adds this as the after-install script.
#
# Chown config and log dirs only on a first install. Upgrades must not reset
# an admin who overrode User=/Group= and chowned those trees to match.

# nFPM: deb $1=configure $2=oldver-or-empty; rpm $1=1 install / $1=2 upgrade.
first_install=0
if [ "${1:-}" = "1" ] || [ "${1:-}" = "install" ]; then
  first_install=1
elif [ "${1:-}" = "configure" ] && [ -z "${2:-}" ]; then
  first_install=1
fi

if [ "${first_install}" = 1 ]; then
  if [ -d /usr/local/etc/notifiarr ]; then
    chown -R notifiarr: /usr/local/etc/notifiarr
  fi
  if [ -d /etc/notifiarr ]; then
    chown -R notifiarr: /etc/notifiarr
  fi
  if [ -d /var/log/notifiarr ]; then
    chown -R notifiarr: /var/log/notifiarr
  fi
  if [ -d /usr/local/var/log/notifiarr ]; then
    chown -R notifiarr: /usr/local/var/log/notifiarr
  fi
fi

if [ -x "/bin/systemctl" ]; then
  # Reload and restart - this starts the application as user nobody.
  /bin/systemctl daemon-reload
  /bin/systemctl enable notifiarr
  /bin/systemctl restart notifiarr
fi
