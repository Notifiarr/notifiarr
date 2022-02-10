NOTIFIARR_API_KEY="abcdef12345"
CHANNEL_ID="940876909158490154"

############
# Use this script on your Unraid server to send notifications to Discord through notifiarr.com.
############
# Setup Unraid:
# * Create the agent directory if needed: mkdir -p /boot/config/plugins/dynamix/notifications/agents/
# * Download this script to /boot/config/plugins/dynamix/notifications/agents/Notifiarr.sh
# * Replace NOTIFIARR_API_KEY at the top of this script with the your Notifiarr.com API key.
# * Replace CHANNEL_ID with the channel you want notices sent to. Right click your channel to get this.
# * In the Unraid webgui, go to Settings -> Notification Settings.
# * Set System Notifications to Enabled.
# * Enable an interval for all notifications you want, and put a checkmark next to Agents for each.
# * If you use the Auto Updater, go into those Settings and Set Notifications to Yes.
############

#### Settings ####

# Debug prints the sent payload into the log file.
DEBUG=false
# greyish, default left sidebar color.
COLOR="28282C"
# yellowish, for any warning.
WARNING_COLOR="918c3d"
# orangish, for any error.
ERROR_COLOR="ff5733"
# bluish, for anything weird.
ABNORMAL_COLOR="383582"
# Format string for `date` (Time).
DATE_FORMAT="+%Y-%m-%d %H:%M:%S"

##### SCRIPT BELOW, CHANGE ONLY IF YOU WANT TO BREAK THINGS #####

# Script name is used for log file name, and logger app.
SCRIPTNAME=$(basename "$0")
LOG="/var/log/notify_${SCRIPTNAME%.*}"
# Sometimes CloudFlare has an error, mitigate with retries.
RETRIES=4
URL="https://notifiarr.com/api/v1/notification/passthrough"

# This stuff may work on OSes besides Unraid, but don't count on it.
#UPTIME_SEC=$(head -n1 /proc/uptime | cut -d. -f1)
UNRAIDVER=$(head -n1 /etc/issue | grep -o '[0-9]\+\.[0-9]\+\(\.[0-9]\+\)\?$')
UPTIME=$(uptime)
LOAD=$(echo "${UPTIME}" | grep -o '[0-9]\+\.[0-9]\+, [0-9]\+\.[0-9]\+, [0-9]\+\.[0-9]\+$')
UPTIME=$(echo "${UPTIME}" | grep -o 'up[^,]\+')
CONTAINERS=$(docker ps -q 2>&1 | wc -l)
KERNEL=$(uname -s -r)

# These values are generally only used when you run the script as a test.
[[ "${EVENT}" ]]       || EVENT='No event'
[[ "${SUBJECT}" ]]     || SUBJECT='No subject'
[[ "${DESCRIPTION}" ]] || DESCRIPTION='No description'
[[ "${IMPORTANCE}" ]]  || IMPORTANCE='test'
[[ "${TIMESTAMP}" ]]   || TIMESTAMP="$(date +%s)"
# Content is optional. Not always present. Add it only if it exists.
[[ ! "${CONTENT}" ]]   || CONTENT=$(cat <<EOF
,
        {"title": "Content", "text": "${CONTENT}"}
EOF
)

# Change the color based on importance.
[ "${IMPORTANCE}" == "normal" ]  || COLOR="${ABNORMAL_COLOR}"
[ "${IMPORTANCE}" != "warning" ] || COLOR="${WARNING_COLOR}"
[ "${IMPORTANCE}" != "error" ]   || COLOR="${ERROR_COLOR}"

# Format the eopch time stamp.
DATETIME=$(date -u "${DATE_FORMAT} UTC" -d @"${TIMESTAMP}")

# Create our payload for the website.
PASSTHRU=$(cat <<EOF
{
  "notification": { "update": false, "name": "Unraid: ${IMPORTANCE^} Alert" },
  "discord": {
    "ids": { "channel": ${CHANNEL_ID} },
    "color": "${COLOR}",
    "text": {
      "description": "${DESCRIPTION}",
      "icon":    "https://raw.githubusercontent.com/limetech/Unraid.net/master/Unraid.net.png",
      "title":   "${SUBJECT}",
      "content": "Unraid: ${IMPORTANCE^} Alert! ${EVENT}",
      "footer":  "${HOSTNAME} • ${KERNEL} v${UNRAIDVER}",
      "fields": [
        {"title": "Event",   "text": "${EVENT}"},
        {"title": "Time",    "text": "${DATETIME}"}${CONTENT},
        {"title": "Load",    "text": "${LOAD}",       "inline": true},
        {"title": "Uptime",  "text": "${UPTIME}",     "inline": true},
        {"title": "Dockers", "text": "${CONTAINERS}", "inline": true}
      ]
    }
  }
}
EOF
)

# Log the payload if debug is enabled.
[ "${DEBUG}" != "true" ] || echo "$(date) Sending Payload to ${URL}: ${PASSTHRU}" >> "${LOG}"

# Retry the curl in case CloudFlare is in a mood.
for ((i = 1; i <= "${RETRIES}"; i++)); do
  output=$(curl -sH "x-api-key: ${NOTIFIARR_API_KEY}" -d "${PASSTHRU}" "${URL}" 2>&1)
  if [ "$?" -eq "0" ]; then
    echo "$(date) Sent notification to notifiarr.com, subject: ${SUBJECT}" >> "${LOG}"
    break
  fi

  # Log the error and curl output.
  echo "$(date) Curl Attempt ${i} of ${RETRIES} failed, subject: ${SUBJECT}" >> "${LOG}"
  echo "$(date) Curl Output: ${output}" >> "${LOG}"

  if [ "${i}" -eq "${RETRIES}" ]; then
    logger -t "${SCRIPTNAME}" "Failed sending notification - retries met"
    break
  fi

  sleep 1
done
