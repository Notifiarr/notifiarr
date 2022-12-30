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
  WGET="curl -sSL"
else
  wget --version > /dev/null 2>&1
  if [ "$?" = "0" ]; then
    WGET="wget -qO-"
  fi
fi

if [ "$WGET" = "" ]; then
  echo "${P} [ERROR] Could not locate curl nor wget - please install one to download files!"
  exit 1
fi

INSTALLED=$(/usr/bin/notifiarr -v 2>/dev/null | head -n1 | cut -d' ' -f3)

# Grab latest release file from github.
PAYLOAD=$($WGET ${LATEST})
URL=$(echo "$PAYLOAD" | egrep "browser_download_url.*(${ARCH})\.${FILE}\"" | cut -d\" -f 4)
TAG=$(echo "$PAYLOAD" | grep 'tag_name' | cut -d\" -f4 | tr -d v)

if [ "$?" != "0" ] || [ "$URL" = "" ]; then
  echo "${P} [ERROR] Missing latest release for '${FILE}' file ($OS/${ARCH}) at ${LATEST}"
  echo "${P} $(uname -a)"
  echo "${P} Please report this error, along with the above OS details:"
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
$WGET ${URL} > /tmp/${FILE}

if [ "$(id -u)" != "0" ]; then
  echo "${P} Downloaded, but no root access. Install the file manually to /usr/bin/notifiarr"
  echo "${P} Recommend re-running this script as root instead!"
  echo "${P} Doing so will install the upstart file, notifiarr user, and config file."
  exit 0
fi

# Install it.
echo "${P} Downloaded. Installing the binary to /usr/bin/notifiarr"

mkdir -p /etc/notifiarr /var/log/notifiarr
gunzip -c /tmp/${FILE} > /usr/bin/notifiarr
rm /tmp/${FILE}
chmod 0755 /usr/bin/notifiarr /var/log/notifiarr
chown notifiarr: /var/log/notifiarr

ID=$(id notifiarr 2>&1)
if [ "$?" != "0" ]; then
  echo "${P} Adding notifiarr user: synouser --add notifiarr Notifiarr 0 support@notifiarr.com 0"
  pass="${RANDOM}${RANDOM}${RANDOM}${RANDOM}${RANDOM}${RANDOM}${RANDOM}${RANDOM}"
  synouser --add notifiarr "${pass}" Notifiarr 0 support@notifiarr.com 0
else
  echo "${P} User notifiarr already exists: ${ID}"
fi

CONFIGFILE=/etc/notifiarr/notifiarr.conf
echo "${P} Ensuring config file: ${CONFIGFILE} and log dir: /var/log/notifiarr"
[ -f "${CONFIGFILE}" ] || $WGET https://docs.notifiarr.com/configs/notifiarr-synology.conf > "${CONFIGFILE}"

echo "${P} Adding sudoers entry to: /etc/sudoers"
sed -i '/notifiarr/d' /etc/sudoers
echo 'notifiarr ALL=(root) NOPASSWD:/bin/smartctl *' >> /etc/sudoers

SYSTEMCTL=$(which systemctl)

if [ -z "$SYSTEMCTL" ]; then
  echo "${P} Updating init file: /usr/share/init/notifiarr.conf"
  cat <<EOT > /usr/share/init/notifiarr.conf
description "start notifiarr"

start on syno.network.ready
stop on runlevel [06]

respawn
respawn limit 5 10

setuid notifiarr
exec /usr/bin/notifiarr -c ${CONFIGFILE}
EOT
else
  echo "${P} Updating unit file: /etc/systemd/system/notifiarr.service"
  cat <<EOT > /etc/systemd/system/notifiarr.service
[Unit]
Description=notifiarr - Official Client for Notifiarr.com
After=network.target
Requires=network.target

[Service]
ExecStart=/usr/bin/notifiarr -c ${CONFIGFILE}
Restart=always
RestartSec=10
SyslogIdentifier=notifiarr
Type=simple
WorkingDirectory=/tmp

[Install]
WantedBy=multi-user.target
EOT
fi

chown -R notifiarr: /etc/notifiarr

echo "${P} Restarting service (if running): status notifiarr ; stop notifiarr ; start notifiarr"
SYSTEMCTL=$(which systemctl)
[ -z "$SYSTEMCTL" ] && status notifiarr || systemctl status notifiarr
if [ "$?" = "0" ]; then
  if [ -z "$SYSTEMCTL" ]; then
    stop notifiarr
    start notifiarr
  else
    systemctl stop notifiarr
    systemctl start notifiarr
  fi
fi

if [ "${INSTALLED}" = "" ]; then
  echo "${P} Installed. Edit your config file: ${CONFIGFILE}"
  echo "${P} Log files are written to: /var/log/notifiarr"
  if [ -z "$SYSTEMCTL" ]; then
    echo "${P} start the service with:  start notifiarr"
    echo "${P} stop the service with:   stop notifiarr"
    echo "${P} to check service status: status notifiarr"
  else
    echo "${P} start the service with:  systemctl notifiarr start"
    echo "${P} stop the service with:   systemtcl notifiarr stop"
    echo "${P} to check service status: systemctl notifiarr status"
  fi
else
  echo "${P} Upgraded and restarted."
fi
echo "<-------------------------------------------------->"
