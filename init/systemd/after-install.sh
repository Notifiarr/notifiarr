#!/bin/sh

# This file is used by deb, rpm and BSD packages.
# FPM/nFPM adds this as the after-install script.
#
# Chown the packaged config files only on a first install. Upgrades must not
# reset an admin who overrode User=/Group= and chowned the config to match.
# Name the packaged files instead of chown -R so extra files in the directory
# keep the ownership the admin set. Only touch this OS's config/log roots.

OS="$(uname -s)"

if [ "${OS}" = "Linux" ]; then
  confdir=/etc/notifiarr
  logdir=/var/log/notifiarr
else
  confdir=/usr/local/etc/notifiarr
  logdir=/usr/local/var/log/notifiarr
fi

# nFPM: deb $1=configure $2=oldver-or-empty; rpm $1=1 install / $1=2 upgrade.
first_install=0
if [ "${1:-}" = "1" ] || [ "${1:-}" = "install" ]; then
  first_install=1
elif [ "${1:-}" = "configure" ] && [ -z "${2:-}" ]; then
  first_install=1
fi

if [ "${first_install}" = 1 ]; then
  if [ -d "${confdir}" ]; then
    chown notifiarr: "${confdir}"
    [ -f "${confdir}/notifiarr.conf" ] && chown notifiarr: "${confdir}/notifiarr.conf"
    [ -f "${confdir}/notifiarr.conf.example" ] && chown notifiarr: "${confdir}/notifiarr.conf.example"
  fi
  if [ -d "${logdir}" ]; then
    chown notifiarr: "${logdir}"
  fi
fi

if [ -x "/bin/systemctl" ]; then
  # Reload and restart - this starts the application as user nobody.
  /bin/systemctl daemon-reload
  /bin/systemctl enable notifiarr
  /bin/systemctl restart notifiarr
fi
