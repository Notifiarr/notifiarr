NOTIFIARR_API_KEY="abcdef12345"
############
# Use this script on your Unraid server to send notifications to Discord through notifiarr.com.
############
# Setup Unraid:
# * Create the agent directory if needed: mkdir -p /boot/config/plugins/dynamix/notifications/agents/
# * Download this script to /boot/config/plugins/dynamix/notifications/agents/Notifiarr.sh
# * Replace NOTIFIARR_API_KEY at the top of this script with the your Notifiarr.com API key.
# * In the Unraid webgui, go to Settings -> Notification Settings.
# * Set System Notifications to Enabled.
# * Enable an interval for all notifications you want, and put a checkmark next to Agents for each.
# * If you use the Auto Updater, go into those Settings and Set Notifications to Yes.
# Setup Notifiarr.com: (IMPORTANT)
# * Enable the Package Manager integration on the Integration Setup page under Manage Integrations.
# * Pick a channel in the same place.
############

##### SCRIPT BELOW, CHANGE ONLY IF YOU WANT TO BREAK THINGS #####

## Script name is used for log file name, and logger app.
SCRIPTNAME=Notifiarr
LOG="/var/log/notify_${SCRIPTNAME%.*}"
## Don't change this.
URL="https://notifiarr.com/api/v1/notification/packageManager?event=unraid"
## Sometimes CloudFlare has an error, mitigate with retries.
RETRIES=4

[[ "${EVENT}" ]]       || EVENT='No event'
[[ "${SUBJECT}" ]]     || SUBJECT='No subject'
[[ "${DESCRIPTION}" ]] || DESCRIPTION='No description'
[[ "${IMPORTANCE}" ]]  || IMPORTANCE='test'
[[ "${TIMESTAMP}" ]]   || TIMESTAMP="$(date +%s)"

UPTIME=$(cut -d. -f1 /proc/uptime)
LOADAV=$(uptime | cut -d: -f5)
KERNEL=$(uname -s -r)
RUNNING=$(docker ps -q 2>&1 | wc -l)
STOPPED=$(echo "$(docker ps -aq 2>&1 | wc -l) ${RUNNING}" | awk '{printf($1-$2)}')
UNRAIDVER=$(head -n1 /etc/issue | grep -o '[0-9]\+\.[0-9]\+\(\.[0-9]\+\)\?$')
MEMTOTAL=$(grep MemTotal /proc/meminfo | awk '{printf($2)}')
MEMAVAIL=$(grep MemAvailable /proc/meminfo | awk '{printf($2)}')
CPUPERC=$(top -b -n2 -p 1 | fgrep "Cpu(s)" | tail -1 | \
  awk -F'id,' -v prefix="$prefix" '{ split($1, vs, ","); v=vs[length(vs)]; sub("%", "", v); printf "%s%.1f", prefix, 100 - v }')

POST=$(cat <<EOF
  {
    "event":   "${EVENT}",
    "subject": "${SUBJECT}",
    "desc":    "${DESCRIPTION}",
    "type":    "${IMPORTANCE}",
    "content": "${CONTENT}",
    "extra": {
      "uptime":   ${UPTIME},
      "running":  ${RUNNING},
      "stopped":  ${STOPPED},
      "cpu":      ${CPUPERC},
      "load":     "${LOADAV}",
      "kernel":   "${KERNEL}",
      "version":  "${UNRAIDVER}",
      "memTotal": "${MEMTOTAL}",
      "memAvail": "${MEMAVAIL}",
      "hostname": "${HOSTNAME}"
    }
  }
EOF
)

[ "${DEBUG}" != "true" ] || echo -e "[$(date)] Sending Payload to ${URL}:\n${POST}" >> "${LOG}"

## Retry the curl in case CloudFlare is in a mood.
for ((retry = 1; retry <= "${RETRIES}"; retry++)); do
  output=$(curl -sH "x-api-key: ${NOTIFIARR_API_KEY}" -d "${POST}" "${URL}" 2>&1)
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
