#!/bin/bash
#
# This is a quick and dirty script to install the latest notifiarr on Synology.
#
# Use it like this, pick curl or wget:  (sudo is not optional for Synology)
# ----
#   curl -sSL https://raw.githubusercontent.com/Notifiarr/notifiarr/main/scripts/install-synology.sh | sudo bash
#   wget -qO- https://raw.githubusercontent.com/Notifiarr/notifiarr/main/scripts/install-synology.sh | sudo bash
# ----
#
# This file can be added to crontab. First, save it locally:
#   sudo curl -sSLo /usr/bin/update-notifiarr.sh https://raw.githubusercontent.com/Notifiarr/notifiarr/main/scripts/install-synology.sh
# Then install this crontab:
#   echo "10 3 * * * root /bin/bash /usr/bin/update-notifiarr.sh 2>&1 > /volume1/data/notifiarr-updates.log" | sudo tee /etc/cron.d/update-notifiarr
#

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

INSTALLED=$(/usr/bin/notifiarr -v 2>/dev/null | head -n1 | cut -d' ' -f3)

# Grab latest release file from github.
PAYLOAD=$($CMD ${LATEST})
URL=$(echo "$PAYLOAD" | egrep "browser_download_url.*(${ARCH})\.${FILE}\"" | cut -d\" -f 4)
TAG=$(echo "$PAYLOAD" | grep 'tag_name' | cut -d\" -f4 | tr -d v)

if [ "$?" != "0" ] || [ "$URL" = "" ]; then
  echo "${P} [ERROR] Missing latest release for '${FILE}' file ($OS/${ARCH}) at ${LATEST}"
  echo "${P} $(uname -a)"
  echo "${P} Please report error this, along with the above OS details:"
  echo "     ${ISSUES}"
  exit 1
fi

# https://stackoverflow.com/questions/4023830/how-to-compare-two-strings-in-dot-separated-version-format-in-bash
vercomp () {
  if [ "$1" = "" ]; then
    return 3
  elif [ "$1" = "$2" ]; then
    return 0
  fi

  local IFS=.
  local i ver1=($1) ver2=($2)
  # fill empty fields in ver1 with zeros
  for ((i=${#ver1[@]}; i<${#ver2[@]}; i++)); do
    ver1[i]=0
  done

  for ((i=0; i<${#ver1[@]}; i++)); do
    if [[ -z ${ver2[i]} ]]; then
      # fill empty fields in ver2 with zeros
      ver2[i]=0
    elif ((10#${ver1[i]} > 10#${ver2[i]})); then
      return 1
    elif ((10#${ver1[i]} < 10#${ver2[i]})); then
      return 2
    fi
  done
  return 0
}

vercomp "$INSTALLED" "$TAG"
case $? in
  0) echo "${P} The installed version of ${PACKAGE} (${INSTALLED}) is current: ${TAG}" ; exit 0 ;;
  1) echo "${P} The installed version of ${PACKAGE} (${INSTALLED}) is newer than the current release (${TAG})" ; exit 0 ;;
  2) echo "${P} Upgrading ${PACKAGE} to ${TAG} from ${INSTALLED}." ;;
  3) echo "${P} Installing ${PACKAGE} version ${TAG}." ;;
esac

FILE=$(basename ${URL})
echo "${P} Downloading: ${URL}"
echo "${P} To Location: /tmp/${FILE}"
$CMD ${URL} > /tmp/${FILE}

if [ "$(id -u)" != "0" ]; then
  echo "${P} Downloaded, but no root access. Install the file manually to /usr/bin/notifiarr"
  echo "${P} Recommend re-running this script as root instead!"
  echo "${P} Doing so will install the upstart file, notifiarr user, and config file."
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
  if [ "${CONFIGFILE}" != "/etc/notifiarr/notifiarr.conf" ]; then
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
