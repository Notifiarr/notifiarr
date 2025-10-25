#!/bin/bash
#
# Use this script to install Notifiarr client on Linux in your home folder (when you don't have root access)
# Can also be used to update the application to the latest release.
#
# Use it like this, pick curl or wget:
# ----
#   curl -sSL https://raw.githubusercontent.com/Notifiarr/notifiarr/main/userscripts/install-local.sh | bash
#   wget -qO- https://raw.githubusercontent.com/Notifiarr/notifiarr/main/userscripts/install-local.sh | bash
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

PACKAGE=notifiarr

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
  echo "${P} [ERROR] Could not locate curl nor wget - make sure one of them is in your PATH. Cannot continue!"
  exit 1
fi

INSTALLED=$($HOME/notifiarr/notifiarr -v 2>/dev/null | cut -d' ' -f 2 | cut -d- -f1)

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

if ! grep -qE 'Environment=DN_API_KEY=[0-9a-fA-F-]{36}' "${HOME}/.config/systemd/user/notifiarr.service" 2>/dev/null; then
  echo -n "${P} Paste your 'All' API Key from notifiarr.com profile page: "
  read API_KEY

  if ! echo "$API_KEY" | grep -q '^[0-9a-fA-F-]\{36\}$'; then
    echo "${P} [ERROR] Invalid API Key format. Must be 36 hexadecimal characters. Cannot continue!"
    exit 1
  fi
fi

mkdir -p "${HOME}/notifiarr"

FILE=$(basename ${URL})
echo "${P} Downloading: ${URL}"
echo "${P} To Location: ${HOME}/notifiarr/${FILE}"
$WGET ${URL} > "${HOME}/notifiarr/${FILE}"

# Install it.
echo "${P} Downloaded. Installing the binary to ${HOME}/notifiarr/notifiarr"
systemctl --user stop notifiarr 2>/dev/null >/dev/null
gunzip -c "${HOME}/notifiarr/${FILE}" > "${HOME}/notifiarr/notifiarr"
rm "${HOME}/notifiarr/${FILE}"
chmod 0755 "${HOME}/notifiarr/notifiarr"

# Create systemd service unit file.
if [ -n "${API_KEY}" ]; then
  echo "${P} Updating unit file: ${HOME}/.config/systemd/user/notifiarr.service"
  echo "${P} Your API key is stored in this file ^^^^ change it there if ever needed."
  mkdir -p "${HOME}/.config/systemd/user"
  cat <<EOT > "${HOME}/.config/systemd/user/notifiarr.service"
# Systemd service unit for notifiarr.

[Unit]
Description=notifiarr - Official chat integration client for Notifiarr.com

[Service]
ExecStart=${HOME}/notifiarr/notifiarr
Restart=always
RestartSec=10
Type=simple
WorkingDirectory=${HOME}/notifiarr
Environment=DN_API_KEY=${API_KEY}
Environment=DN_URLBASE=/notifiarr
Environment=DN_LOG_FILE=${HOME}/notifiarr/app.log
Environment=DN_HTTP_LOG=${HOME}/notifiarr/http.log
Environment=DN_DEBUG_LOG=${HOME}/notifiarr/debug.log
Environment=DN_SERVICES_LOG_FILE=${HOME}/notifiarr/services.log
Environment=DN_QUIET=true

[Install]
WantedBy=default.target
EOT
fi

systemctl --user daemon-reload
systemctl --user enable notifiarr
systemctl --user start notifiarr

# Create nginx proxy configuration for the Notifiarr Web UI if the folder exists and the file does not.
if [ -d "${HOME}/.apps/nginx/proxy.d" ] && [ ! -f "${HOME}/.apps/nginx/proxy.d/notifiarr.conf" ]; then
  echo "${P} Creating nginx proxy configuration for the Notifiarr Web UI."
  cat <<EOT > "${HOME}/.apps/nginx/proxy.d/notifiarr.conf"
location /notifiarr {
    # <put proxy auth directives here> Optional:
    # proxy_set_header X-WebAuth-User \$auth_user;
    proxy_set_header X-Forwarded-For \$remote_addr;
    set \$notifiarr http://127.0.0.1:5454;
    proxy_pass \$notifiarr\$request_uri;
    proxy_http_version 1.1;
    proxy_set_header Upgrade \$http_upgrade;
    proxy_set_header Connection \$http_connection;
    proxy_set_header Host \$host;
}
# Notifiarr Client
location /notifiarr/api {
    proxy_set_header X-Forwarded-For \$remote_addr;
    set \$notifiarr http://127.0.0.1:5454;
    proxy_pass \$notifiarr\$request_uri;
}
EOT

  if systemctl --user is-active nginx >/dev/null; then systemctl --user restart nginx; fi
fi

if [ "${INSTALLED}" = "" ]; then
  echo "${P} Installed! Open the Web UI @ http://your-local-or-seedbox-url:5454 and login as 'admin' using your API Key as the password."
  [ -f "${HOME}/.apps/nginx/proxy.d/notifiarr.conf" ] &&
    echo "${P} If nginx is running you should be able to access the Web UI at https://your-local-or-seedbox-url/notifiarr"
  echo "${P} Start the service with:  systemctl --user start notifiarr"
  echo "${P} Stop the service with:   systemctl --user stop notifiarr"
  echo "${P} Check service status:    systemctl --user status notifiarr"
else
  echo "${P} Upgraded to ${TAG} and restarted."
fi

echo "<-------------------------------------------------->"
