#!/bin/bash
#
# This is a quick and dirty script to install the latest notifiarr on Synology.
#
# Use it like this, pick curl or wget:  (sudo is optional)
# ----
#   curl -sL https://raw.githubusercontent.com/Notifiarr/notifiarr/main/scripts/install-synology.sh | sudo bash
#   wget -qO- https://raw.githubusercontent.com/Notifiarr/notifiarr/main/scripts/install-synology.sh | sudo bash
# ----

# Set the repo name correctly.
REPO=Notifiarr/notifiarr

# Nothing else needs to be changed. Unless you're fixing things!

LATEST=https://api.github.com/repos/${REPO}/releases/latest
ISSUES=https://github.com/${REPO}/issues/new
ARCH=$(uname -m)
OS=$(uname -s)
P=" ==>"

echo "<-------------------------------------------------->"

# $ARCH is passed into egrep to find the right file.
if [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "amd64" ]; then
  ARCH="x86_64|amd64"
elif [[ $ARCH = *386* ]] || [[ $ARCH = *686* ]]; then
  ARCH="i386"
elif [[ $ARCH = *arm64* ]] || [[ $ARCH = *armv8* ]] || [[ $ARCH = *aarch64* ]]; then
  ARCH="arm64"
elif [[ $ARCH = *armv6* ]] || [[ $ARCH = *armv7* ]]; then
  ARCH="armhf"
else
  echo "${P} [ERROR] Unknown Architecture: ${ARCH}"
  echo "${P} $(uname -a)"
  echo "${P} Please report this, along with the above OS details:"
  echo "     ${ISSUES}"
  exit 1
fi

FILE=linux.gz

# curl or wget?
curl --version > /dev/null 2>&1
if [ "$?" = "0" ]; then
  CMD="curl -sL"
else
  wget --version > /dev/null 2>&1
  if [ "$?" = "0" ]; then
    CMD="wget -qO-"
  fi
fi

if [ "$CMD" = "" ]; then
  echo "${P} [ERROR] Could not locate curl nor wget - please install one to download files!"
  exit 1
fi

# Grab latest release file from github.
URL=$($CMD ${LATEST} | egrep "browser_download_url.*(${ARCH})\.${FILE}\"" | cut -d\" -f 4)

if [ "$?" != "0" ] || [ "$URL" = "" ]; then
  echo "${P} [ERROR] Missing latest release for '${FILE}' file ($OS/${ARCH}) at ${LATEST}"
  echo "${P} $(uname -a)"
  echo "${P} Please report error this, along with the above OS details:"
  echo "     ${ISSUES}"
  exit 1
fi

FILE=$(basename ${URL})
echo "${P} Downloading: ${URL}"
echo "${P} To Location: /tmp/${FILE}"
$CMD ${URL} > /tmp/${FILE}

if [ "$(id -u)" != "0" ]; then
  echo "${P} Downloaded. Install the package manually."
  exit 0
fi

# Install it.
echo "${P} Downloaded. Installing the binary to /usr/bin/notifiarr"

gunzip -c /tmp/${FILE} > /usr/bin/notifiarr
rm /tmp/${FILE}
chmod 0755 /usr/bin/notifiarr

echo "${P} Ensuring config file: /etc/notifiarr/notifiarr.conf (File exists error after this is OK!)"
mkdir /etc/notifiarr 2>/dev/null || \
  $CMD https://notifiarr.com/scripts/notifiarr-client.conf > /etc/notifiarr/notifiarr.conf

CONFIGFILE=/etc/notifiarr/notifiarr.conf
if [ -d /volume1/data ]; then
  CONFIGFILE=/volume1/data/notifiarr.conf
  [ -f /volume1/data/otifiarr.conf ] || cp /etc/notifiarr/notifiarr.conf ${CONFIGFILE}
fi

echo "${P} Adding sudoers entry to: /etc/sudoers"
sed -i '/notifiarr/d' /etc/sudoers
echo 'notifiarr ALL=(root) NOPASSWD:/bin/smartctl *' >> /etc/sudoers

echo "${P} Updating init file: /usr/share/init/notifiarr.conf"
cat <<EOT > /usr/share/init/notifiarr.conf
description "start notifiarr"

start on syno.network.ready
stop on runlevel [06]

respawn
respawn limit 5 10

setuid notifiarr
exec /usr/bin/notifiarr -c /volume1/data/notifiarr.conf
EOT

ID=$(id notifiarr 2>&1)
if [ "$?" != "0" ]; then
  echo "${P} Adding notifiarr user: synouser --add notifiarr Notifiarr 0 support@notifiarr.com 0"
  pass="${RANDOM}${RANDOM}${RANDOM}${RANDOM}${RANDOM}${RANDOM}${RANDOM}${RANDOM}"
  synouser --add notifiarr "${pass}" Notifiarr 0 support@notifiarr.com 0
  if [ "${CONFIGFILE}" != "/etc/notifiarr/notifiarr.conf" ] then
    echo "${P} Authorizing notifiarr user: synoacltool -add /volume1/data user:notifiarr:allow:r--------------:fd--"
    synoacltool -add /volume1/data user:notifiarr:allow:r--------------:fd--
  fi
else
  echo "${P} User notifiarr already exists: ${ID}"
fi

echo "${P} Restarting service (if running): status notifiarr ; stop notifiarr ; start notifiarr"
status notifiarr
if [ "$?" = "0" ]; then
  stop notifiarr
  start notifiarr
fi

echo "${P} Installed. Edit your config file: ${CONFIGFILE}"
echo "${P} start the service with:  start notifiarr"
echo "${P} stop the service with:   stop notifiarr"
echo "${P} to check service status: status notifiarr"
echo "<-------------------------------------------------->"
