#!/bin/sh

# This file is used by deb, rpm and BSD packages.
# FPM adds this as the before-install script.

OS="$(uname -s)"

if [ "${OS}" = "Linux" ]; then
  # Make a user and group for this app, but only if it does not already exist.
  id notifiarr >/dev/null 2>&1  || \
    useradd --system --user-group --no-create-home --home-dir /tmp --shell /bin/false notifiarr
elif [ "${OS}" = "OpenBSD" ]; then
  id notifiarr >/dev/null 2>&1  || \
    useradd  -g =uid -d /tmp -s /bin/false notifiarr
elif [ "${OS}" = "FreeBSD" ]; then
  id notifiarr >/dev/null 2>&1  || \
    pw useradd notifiarr -d /tmp -w no -s /bin/false
else
  echo "Unknown OS: ${OS}, please add system user notifiarr manually."
fi