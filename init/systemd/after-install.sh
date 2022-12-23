#!/bin/sh

# This file is used by deb, rpm and BSD packages.
# FPM adds this as the after-install script.

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

if [ -x "/bin/systemctl" ]; then
  # Reload and restart - this starts the application as user nobody.
  /bin/systemctl daemon-reload
  /bin/systemctl enable notifiarr
  /bin/systemctl restart notifiarr
fi
