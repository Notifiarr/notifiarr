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
# Setup Notifiarr.com:
# * Enable the passthrough integration on the Integration Setup page under Manage Integrations.
############

#### More Settings ####

## greyish, normal event left sidebar color.
COLOR="28282C"
## yellowish, for any warning.
WARNING_COLOR="918c3d"
## orangish, for any error.
ERROR_COLOR="ff5733"
## bluish, for anything weird (unlikely), and tests.
ABNORMAL_COLOR="383582"
## Format string for `date` (Time). Always in UTC.
DATE_FORMAT="+%Y-%m-%d %H:%M:%S"
## You can change this if you really want to.
UNRAID_LOGO="https://craftassets.unraid.net/uploads/logos/un-mark-gradient@2x.png"
## Debug prints the sent payload into the log file: /var/log/notify_Notifiarr
## Remove comment (#) on the next line to enable debug.
#DEBUG=true

##### SCRIPT BELOW, CHANGE ONLY IF YOU WANT TO BREAK THINGS #####

## Script name is used for log file name, and logger app.
SCRIPTNAME=Notifiarr
LOG="/var/log/notify_${SCRIPTNAME%.*}"
## Don't change this.
URL="https://notifiarr.com/api/v1/notification/passthrough"
## Sometimes CloudFlare has an error, mitigate with retries.
RETRIES=4

## This stuff may work on OSes besides Unraid, but don't count on it.
UPTIME=$(uptime -p | sed -e 's/^up//g' -e 's/[aeiou]//g')
LOADAV=$(uptime | cut -d: -f5)
KERNEL=$(uname -s -r)
DOCKER=$(docker ps -q 2>&1 | wc -l)
STOPED=$(docker ps -aq 2>&1 | wc -l)
STOPED=$(echo "${STOPED} ${DOCKER}" | awk '{printf($1-$2)}')
UNRAIDVER=$(head -n1 /etc/issue | grep -o '[0-9]\+\.[0-9]\+\(\.[0-9]\+\)\?$')
MEMTOTAL=$(grep MemTotal /proc/meminfo | awk '{printf($2/1024)}')
MEMAVAIL=$(grep MemAvailable /proc/meminfo | awk '{printf($2/1024)}')
MEMUSED=$(echo "${MEMTOTAL} ${MEMAVAIL}" | awk '{printf("%.1f", $1-$2)}')
MEMPERC=$(echo "${MEMTOTAL} ${MEMUSED}" | awk '{printf("%.1f", $2/$1*100.0)}')
CPUPERC=$(top -b -n2 -p 1 | fgrep "Cpu(s)" | tail -1 | \
  awk -F'id,' -v prefix="$prefix" '{ split($1, vs, ","); v=vs[length(vs)]; sub("%", "", v); printf "%s%.1f", prefix, 100 - v }')

# These values are generally only used when you run the script as a test.
[[ "${EVENT}" ]]       || EVENT='No event'
[[ "${SUBJECT}" ]]     || SUBJECT='No subject'
[[ "${DESCRIPTION}" ]] || DESCRIPTION='No description'
[[ "${IMPORTANCE}" ]]  || IMPORTANCE='test'
[[ "${TIMESTAMP}" ]]   || TIMESTAMP="$(date +%s)"
## Content is optional. Not always present. Add it only if it exists.
[[ ! "${CONTENT}" ]]   || CONTENT=$(cat <<EOF
,
        {"title": "Content", "text": "${CONTENT}"}
EOF
)

## Change the color based on importance.
[ "${IMPORTANCE}" == "normal" ]  || COLOR="${ABNORMAL_COLOR}"
[ "${IMPORTANCE}" != "warning" ] || COLOR="${WARNING_COLOR}"
[ "${IMPORTANCE}" != "error" ]   || COLOR="${ERROR_COLOR}"

## Format the eopch time stamp.
DATETIME=$(date -u "${DATE_FORMAT}" -d @"${TIMESTAMP}")

## Create our payload for the passthrough integration on the website.
PASSTHRU=$(cat <<EOF
{
  "notification": { "update": false, "name": "Unraid: ${IMPORTANCE^} Alert" },
  "discord": {
    "ids": { "channel": ${CHANNEL_ID} },
    "color": "${COLOR}",
    "text": {
      "description": "${EVENT}\n${DESCRIPTION}",
      "icon":    "${UNRAID_LOGO}",
      "title":   "${SUBJECT}",
      "content": "Unraid: ${IMPORTANCE^} Alert! ${EVENT}",
      "footer":  "${HOSTNAME} • ${KERNEL} v${UNRAIDVER}",
      "fields": [
        {"title": "UTC",     "text": "${DATETIME}"}${CONTENT},
        {"title": "Load",    "text": "${LOADAV}",   "inline": true},
        {"title": "CPU",     "text": "${CPUPERC}%", "inline": true},
        {"title": "RAM",     "text": "${MEMPERC}% [${MEMUSED}MB]", inline:true},
        {"title": "Uptime",  "text": "${UPTIME}", "inline": true},
        {"title": "Dockers", "text": "${DOCKER}", "inline": true},
        {"title": "Stopped", "text": "${STOPED}", "inline": true}
      ]
    }
  }
}
EOF
)

## Log the payload if debug is enabled.
[ "${DEBUG}" != "true" ] || echo "[$(date)] Sending Payload to ${URL}:\n${PASSTHRU}" >> "${LOG}"

## Retry the curl in case CloudFlare is in a mood.
for ((retry = 1; retry <= "${RETRIES}"; retry++)); do
  output=$(curl -sH "x-api-key: ${NOTIFIARR_API_KEY}" -d "${PASSTHRU}" "${URL}" 2>&1)
  if [ "$?" -eq "0" ]; then
    echo "[$(date)] Sent notification to notifiarr.com, subject: ${SUBJECT}" >> "${LOG}"
    break
  fi

  ## Log the error and curl output.
  echo "[$(date)] Curl Attempt ${retry} of ${RETRIES} failed, subject: ${SUBJECT}" >> "${LOG}"
  echo "[$(date)] Curl Output: ${output}" >> "${LOG}"

  if [ "${retry}" -eq "${RETRIES}" ]; then
    echo "[$(date)] Curl Failed, giving up after ${RETRIES} retries." >> "${LOG}"
    logger -t "${SCRIPTNAME}" "Failed sending notification - retries met"
    break
  fi

  sleep 1
done
