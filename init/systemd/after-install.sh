#!/bin/sh

# This file is used by deb, rpm, nFPM archlinux, and FreeBSD packages.
#
# Claim packaged config paths that are still root-owned. Do not use packager
# $1/$2 as "first install": nFPM Arch passes the version string, and FreeBSD
# pkg runs this script with no args on install *and* upgrade. An admin who
# overrode User=/Group= has already chowned away from root, so we leave them
# alone. Extra files in the directory are not touched.

OS="$(uname -s)"

if [ "${OS}" = "Linux" ]; then
  confdir=/etc/notifiarr
  logdir=/var/log/notifiarr
else
  confdir=/usr/local/etc/notifiarr
  logdir=/usr/local/var/log/notifiarr
fi

chown_if_root() {
  [ -e "$1" ] || return 0
  uid="$(stat -c '%u' "$1" 2>/dev/null || stat -f '%u' "$1")"
  [ "${uid}" = 0 ] || return 0
  chown notifiarr: "$1"
}

chown_if_root "${confdir}"
chown_if_root "${confdir}/notifiarr.conf"
chown_if_root "${confdir}/notifiarr.conf.example"

if [ ! -d "${logdir}" ]; then
  mkdir "${logdir}"
  chown notifiarr: "${logdir}"
  chmod 0755 "${logdir}"
else
  chown_if_root "${logdir}"
fi

if [ -x "/bin/systemctl" ]; then
  # Reload and restart - this starts the application as user nobody.
  /bin/systemctl daemon-reload
  /bin/systemctl enable notifiarr
  /bin/systemctl restart notifiarr
fi
