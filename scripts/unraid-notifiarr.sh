NOTIFIARR_API_KEY=""
CHANNEL_ID="940876909158490154"

############
# Use this script on your unRaid server to send notifications to notifiarr.com.
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

##### SCRIPT BELOW #####

# Available fields from notification system
# HOSTNAME
# EVENT (notify -e)
# IMPORTANCE (notify -i)
# SUBJECT (notify -s)
# DESCRIPTION (notify -d)
# CONTENT (notify -m)
# TIMESTAMP (seconds from epoch)

UNRAIDVER=$(head -n1 /etc/issue | grep -o '[0-9]\+\.[0-9]\+\(\.[0-9]\+\)\?$')
#UPTIME_SEC=$(head -n1 /proc/uptime | cut -d. -f1)
UPTIME=$(uptime | grep -o 'up[^,]\+')
LOAD=$(uptime | grep -o '[0-9]\+\.[0-9]\+, [0-9]\+\.[0-9]\+, [0-9]\+\.[0-9]\+$')
CONTAINERS=$(docker ps | wc -l)

# for quick test, setup environment to mimic notify script
[[ "${EVENT}" ]]       || EVENT='No event'
[[ "${SUBJECT}" ]]     || SUBJECT='No subject'
[[ "${DESCRIPTION}" ]] || DESCRIPTION='No description'
[[ "${IMPORTANCE}" ]]  || IMPORTANCE='No importance'
[[ "${TIMESTAMP}" ]]   || TIMESTAMP=$(date +%s)
[[ ! "${CONTENT}" ]]   || CONTENT=$(cat <<EOF
, {"title": "Content", "text": "${CONTENT}"}
EOF
)

COLOR="000000"
[ "${IMPORTANCE}" == "normal" ] || COLOR="FF5733"

PASSTHRU=$(cat <<EOF
{
    "notification": {
        "update": false,
        "name": "${HOSTNAME}",
        "event": ""
    },
    "discord": {
        "color": "${COLOR}",
        "images": {
            "thumbnail": "https://raw.githubusercontent.com/limetech/Unraid.net/master/Unraid.net.png"
        },
        "text": {
            "title": "${EVENT}: ${SUBJECT}",
            "content": "unRAID Alert! ${IMPORTANCE}: ${EVENT}",
            "footer": "unRAID v${UNRAIDVER}",
            "fields": [
              {"title": "Description", "text": "${DESCRIPTION}"}${CONTENT},
              {"title": "Load Avg", "text": "${LOAD}", "inline": true},
              {"title": "Uptime", "text": "${UPTIME}", "inline": true},
              {"title": "Dockers", "text": "${CONTAINERS}", "inline": true}
            ]
        },
        "ids": {
            "channel": ${CHANNEL_ID}
        }
    }
}
EOF
)

curl -sH "x-api-key: ${NOTIFIARR_API_KEY}" -d "${PASSTHRU}" \
        "https://dev.notifiarr.com/api/v1/notification/passthrough?event=unraid" 2>&1 > /dev/null
