#!/bin/bash

####################################################################################################
# Use this script to keep your Synology up to date with unstable releases.                         #
# Please only use this script if you're an active Notifiarr tester or on the support team.         #
# Regular users (non-testers) should use the install-synology.sh script also found in this folder. #
####################################################################################################

set -e

# This is where the log file goes, and where temporary downloads are stored.
# It must exist, and be writable by the user running this script.
WORKDIR="/volume1/homes/notifiarr"
# Unstable website URL. This redirects to golift.io/notifiarr/.
UNSTABLE="https://unstable.notifiarr.app"
# File to download/check for updates. Do not include the .gz or .txt suffixes.
FILE="notifiarr.amd64.linux"
# Location for script output.
LOG_FILE="${WORKDIR}/update-log-notifiarr.txt"

####################################################################################################

LATEST="notifiarr $(wget -qO- ${UNSTABLE}/${FILE}.gz.txt | jq -r '"\(.version)-\(.revision)"')"
INSTALLED=$(/usr/bin/notifiarr -v)
TIMESTAMP() { date "+%Y-%m-%d %H:%M:%S" ;}

echo "`TIMESTAMP` latest: ${LATEST}, installed: ${INSTALLED}" >> "${LOG_FILE}"

# Check if current equals unstable.
if [ "$LATEST" = "$INSTALLED" ]; then 
    echo "`TIMESTAMP` Current, exiting. ${LATEST} == ${INSTALLED}" >> "${LOG_FILE}"
    exit 0
fi

echo "`TIMESTAMP` Updating to ${LATEST} from ${INSTALLED}" >> "${LOG_FILE}"

# Download current unstable binary.
wget -qO "${WORKDIR}/${FILE}.gz" ${UNSTABLE}/${FILE}.gz >> "${LOG_FILE}" 2>&1

# Decompress and make executable.
gunzip -f "${WORKDIR}/${FILE}.gz" >> "${LOG_FILE}" 2>&1
chmod 0755 "${WORKDIR}/${FILE}" >> "${LOG_FILE}" 2>&1

# Stop app, move binary into place, start app.
# sudo is only needed if this is not running as root.
sudo systemctl stop notifiarr >> "${LOG_FILE}" 2>&1
sudo mv -f "${WORKDIR}/${FILE}" /usr/bin/notifiarr >> "${LOG_FILE}" 2>&1
sudo systemctl start notifiarr >> "${LOG_FILE}" 2>&1
