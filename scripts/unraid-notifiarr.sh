NOTIFIARR_API_KEY=""
CHANNEL_ID="940876909158490154"

############
# Use this script on your Unraid server to send notifications to Discord through notifiarr.com.
############
# Setup Unraid:
# * Create the agent directory if needed: mkdir -p /boot/config/plugins/dynamix/notifications/agents/
# * Download this script to /boot/config/plugins/dynamix/notifications/agents/Notifiarr.sh
# * Make it executable: chmod +x /boot/config/plugins/dynamix/notifications/agents/Notifiarr.sh
# * Replace NOTIFIARR_API_KEY at the top of this script with the your Notifiarr.com API key.
# * Replace CHANNEL_ID with the channel you want notices sent to. Right click your channel to get this.
# * In the Unraid webgui, go to Settings -> Notification Settings.
# * Set System Notifications to Enabled.
# * Enable an interval for all notifications you want, and put a checkmark next to Agents for each.
# * If you use the Auto Updater, go into those Settings and Set Notifications to Yes.
#
# Inspiration: https://gist.github.com/ljm42/fa80ad36c20e9ae0b158b714865f67e2
#

# greyish, default left sidebar color.
COLOR="28282C"
# orangish, any importance != "normal" alert.
ABNORMAL_COLOR="FF5733"

##### SCRIPT BELOW, CHANGE ONLY IF YOU WANT TO BREAK THINGS #####

# This stuff may work on OSes besides Unraid, but don't count on it.
#UPTIME_SEC=$(head -n1 /proc/uptime | cut -d. -f1)
UNRAIDVER=$(head -n1 /etc/issue | grep -o '[0-9]\+\.[0-9]\+\(\.[0-9]\+\)\?$')
UPTIME=$(uptime)
LOAD=$(echo "${UPTIME}" | grep -o '[0-9]\+\.[0-9]\+, [0-9]\+\.[0-9]\+, [0-9]\+\.[0-9]\+$')
UPTIME=$(echo "${UPTIME}" | grep -o 'up[^,]\+')
CONTAINERS=$(docker ps -q 2>&1 | wc -l)
KERNEL=$(uname -s -r)

# Available fields from notification system
# HOSTNAME
# EVENT (notify -e)
# IMPORTANCE (notify -i)
# SUBJECT (notify -s)
# DESCRIPTION (notify -d)
# CONTENT (notify -m)
# TIMESTAMP (seconds from epoch)

# These values are generally only used when you run the script as a test.
[[ "${EVENT}" ]]       || EVENT='No event'
[[ "${SUBJECT}" ]]     || SUBJECT='No subject'
[[ "${DESCRIPTION}" ]] || DESCRIPTION='No description'
[[ "${IMPORTANCE}" ]]  || IMPORTANCE='abnormal'
[[ ! "${CONTENT}" ]]   || CONTENT=$(cat <<EOF
,
{"title": "Content", "text": "${CONTENT}"}
EOF
)

# Change the color based on importance.
[ "${IMPORTANCE}" == "normal" ] || COLOR="${ABNORMAL_COLOR}"

PASSTHRU=$(cat <<EOF
{
  "notification": { "update": false, "name": "${EVENT}" },
  "discord": {
    "ids": { "channel": ${CHANNEL_ID} },
    "color": "${COLOR}",
    "text": {
        "icon": "https://raw.githubusercontent.com/limetech/Unraid.net/master/Unraid.net.png",
       "title": "${SUBJECT}",
     "content": "${IMPORTANCE^} Unraid Alert! ${EVENT}",
      "footer": "${HOSTNAME} • ${KERNEL} v${UNRAIDVER}",
      "fields": [
        {"title": "Description", "text": "${DESCRIPTION}"}${CONTENT},
        {"title": "Load Avg",    "text": "${LOAD}",       "inline": true},
        {"title": "Uptime",      "text": "${UPTIME}",     "inline": true},
        {"title": "Dockers",     "text": "${CONTAINERS}", "inline": true}
      ]
    }
  }
}
EOF
)

curl -sH "x-api-key: ${NOTIFIARR_API_KEY}" -d "${PASSTHRU}" \
        "https://notifiarr.com/api/v1/notification/passthrough?event=unraid" 2>&1 > /dev/null
