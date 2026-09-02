#!/bin/sh

# This file is used by deb, rpm, nFPM archlinux, and FreeBSD packages.
#
# Claim packaged paths that are still root:root. Do not use packager $1/$2 as
# "first install": nFPM Arch passes the version string, and FreeBSD pkg runs
# this script with no args on install *and* upgrade.
#
# uid 0 alone is not "unclaimed": a User=/Group= override is often
# root:www-data with 0750/0640. Require uid and gid 0. Skip symlinks so a
# root-owned link cannot move the chown onto a file outside the config dir.

OS="$(uname -s)"

if [ "${OS}" = "Linux" ]; then
  confdir=/etc/notifiarr
  logdir=/var/log/notifiarr
else
  confdir=/usr/local/etc/notifiarr
  logdir=/usr/local/var/log/notifiarr
fi

chown_if_root_root() {
  [ -L "$1" ] && return 0
  [ -e "$1" ] || return 0
  ug="$(stat -c '%u %g' "$1" 2>/dev/null || stat -f '%u %g' "$1")"
  [ "${ug}" = "0 0" ] || return 0
  chown notifiarr: "$1"
}

chown_if_root_root "${confdir}"
chown_if_root_root "${confdir}/notifiarr.conf"
chown_if_root_root "${confdir}/notifiarr.conf.example"

if [ ! -d "${logdir}" ]; then
  mkdir "${logdir}"
  chown notifiarr: "${logdir}"
  chmod 0755 "${logdir}"
else
  chown_if_root_root "${logdir}"
fi

if [ -x "/bin/systemctl" ]; then
  # Reload and restart - this starts the application as user nobody.
  /bin/systemctl daemon-reload
  /bin/systemctl enable notifiarr
  /bin/systemctl restart notifiarr
fi
