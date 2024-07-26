#!/bin/sh
#
# FreeBSD rc.d startup script for notifiarr.
#
# PROVIDE: notifiarr
# REQUIRE: networking syslog
# KEYWORD:

. /etc/rc.subr

name="notifiarr"
real_name="notifiarr"
rcvar="notifiarr_enable"
notifiarr_command="/usr/local/bin/${real_name}"
notifiarr_user="notifiarr"
notifiarr_config="/usr/local/etc/${real_name}/notifiarr.conf"
pidfile="/var/run/${real_name}/pid"

# This runs `daemon` as the `notifiarr_user` user.
command="/usr/sbin/daemon"
command_args="-P ${pidfile} -r -t ${real_name} -T ${real_name} -l daemon ${notifiarr_command} -c ${notifiarr_config}"

load_rc_config ${name}
: ${notifiarr_enable:=no}

# Make a place for the pid file.
mkdir -p $(dirname ${pidfile})
chown -R $notifiarr_user $(dirname ${pidfile})

# Suck in optional exported override variables.
# ie. add something like the following to this file: export DN_DEBUG=true
[ -f "/usr/local/etc/defaults/${real_name}" ] && . "/usr/local/etc/defaults/${real_name}"

export DN_LOG_FILE=/usr/local/var/log/notifiarr/app.log
export DN_HTTP_LOG=/usr/local/var/log/notifiarr/http.log
export DN_SERVICES_LOG_FILE=/usr/local/var/log/notifiarr/services.log
export DN_QUIET=true

# Go!
run_rc_command "$1"
