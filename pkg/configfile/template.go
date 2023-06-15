package configfile

import (
	"fmt"
	"runtime"
	"strings"
	"text/template"
	"time"

	"golift.io/version"
)

//nolint:gochecknoglobals
var (
	// ForceAllTmpl allows you to force some specific settings. Used to build a default template.
	ForceAllTmpl = false

	// Template is the config file template.
	Template = template.Must(template.New("config").Funcs(Funcs()).Parse(tmpl))
)

// Funcs returns our template functions.
func Funcs() template.FuncMap {
	return map[string]interface{}{
		"os":    func() string { return runtime.GOOS },
		"force": func() bool { return ForceAllTmpl },
		"toml": func(v string) string {
			return strings.ReplaceAll(v, "'''", `''`)
		},
		"version": func() string {
			return fmt.Sprintf("v%-7s @ %s", version.Version, time.Now().UTC().Format("060201T1504"))
		},
	}
}

//nolint:lll
const tmpl = `###############################################
# Notifiarr Client Example Configuration File #
# Created by Notifiarr {{version}} #
###############################################

## This API key must be copied from your notifiarr.com account.
{{if .APIKey}}api_key = "{{.APIKey}}"{{else}}api_key = "api-key-from-notifiarr.com"{{end}}{{if .ExKeys}}
extra_keys = [{{range $s := .ExKeys}}"{{$s}}",{{end}}]{{end}}

## Setting a UI password properly secures the Web UI. Must be at least 9 characters.
## The default username is admin; change it by setting ui_password to "username:password"
## Set to "webauth" to disable the login form and use only proxy authentication. See upstreams, below.
## Your auth proxy must pass the x-webauth-user header if you set this to "webauth".
## You may also set a custom auth header by setting to "webauth:<header>" e.g. "webauth:remote-user"
## Disable auth by setting this to "noauth". Not recommended and requires "upstreams" being set.
## If you leave this unset the password will be set to the API key defined at the top of this file.
## Changing the password in the Web UI encrypts it, and that is recommended.
ui_password = "{{.UIPassword}}"

## The ip:port to listen on for incoming HTTP requests. 0.0.0.0 means all/any IP and is recommended!
## You may use "127.0.0.1:5454" to listen only on localhost; good if using a local proxy.
## This is used to receive Plex webhooks and Media Request commands.
##
bind_addr = "{{.BindAddr}}"

{{- if or (eq os "windows") (force)}}

## This application can update itself on Windows systems.
## Set this to "daily" to check GitHub every day for updates.
## You may also set it to a Go duration like "12h" or "72h".
## THIS ONLY WORKS ON WINDOWS
{{if .AutoUpdate}}auto_update = "{{.AutoUpdate}}"{{else}}auto_update = "off"{{end}}
{{- end}}

## Quiet makes the app not log anything to output.
## Recommend setting log files if you make the app quiet.
## This is always true on Windows and macOS app.
## Log files are automatically written on those platforms.
##
quiet = {{.Quiet}}

## Debug prints more data and json payloads. This increases application memory usage.
debug = {{.Debug}}
max_body = {{.MaxBody}} # maximum body size for debug logs. 0 = no limit.

{{- if not force}}

## HostID is used to uniquely identify this client's system.
## Do not set or change this unless you have a good reason.
## Do not copy this value to another server, every client must be unique.
host_id = "{{.HostID}}"{{end}}

## All API paths start with /api. This does not affect incoming /plex webhooks.
## Change it to /somethingelse/api by setting urlbase to "/somethingelse"
##
urlbase = "{{.URLBase}}"

## Allowed upstream networks. Networks here are allowed to send two special headers:
## (1) x-forwarded-for (2) x-webauth-user
## The first header sets the IPs in logs.
## The second header allows an auth proxy to set a logged-in username. Be careful.
##
## Set this to your reverse proxy server's IP or network. If you leave off the mask,
## then /32 or /128 is assumed depending on IP version. Empty by default. Example:
##
{{if .Upstreams }}upstreams = [{{range $s := .Upstreams}}"{{$s}}",{{end}}]
{{- else}}#upstreams = [ "127.0.0.1/32", "::1/128" ]{{end}}

## If you provide a cert and key file (pem) paths, this app will listen with SSL/TLS.
## Uncomment both lines and add valid file paths. Make sure this app can read them.
##
{{if .SSLKeyFile}}ssl_key_file  = '{{.SSLKeyFile}}'{{else}}#ssl_key_file  = '/path/to/cert.key'{{end}}
{{if .SSLCrtFile}}ssl_cert_file = '{{.SSLCrtFile}}'{{else}}#ssl_cert_file = '/path/to/cert.key'{{end}}

## If you set these, logs will be written to these files.
## If blank on windows or macOS, log file paths are chosen for you.
{{if .LogFile}}log_file = '{{.LogFile}}'{{else}}#log_file = '~/.notifiarr/notifiarr.log'{{end}}
{{if .HTTPLog}}http_log = '{{.HTTPLog}}'{{else}}#http_log = '~/.notifiarr/notifiarr.http.log'{{end}}{{if or .LogConfig.DebugLog .Debug}}
##
## Debug Log is optional. By default, debug logs write to the app log (above).
## Change that by setting a debug log file path here.
{{if .LogConfig.DebugLog}}debug_log = '{{.LogConfig.DebugLog}}'{{else}}#debug_log = '~/.notifiarr/debug.log'{{end}}{{end}}
##
## Set this to the number of megabytes to rotate files.
log_file_mb = {{.LogFileMb}}
##
## How many files to keep? 0 = all.
log_files = {{.LogFiles}}
##
## Unix file mode for new log files. Umask also affects this.
## Missing, blank or 0 uses default of 0600. Permissive is 0644. Ignored by Windows.
file_mode = "{{.FileMode.String}}"

## Web server and website timeout.
##
timeout = "{{.Timeout}}"

{{- if or (eq os "linux") (force)}}

## This application can integrate with apt on Debian-based OSes.
## Set apt to true to enable this integration. A true setting causes
## notifiarr to relay apt package install/update hooks to notifiarr.com.
##
apt = {{.EnableApt}}
{{- end}}

## Setting serial to true makes the app use fewer threads when polling apps.
## This spreads CPU usage out and uses a bit less memory.
serial = {{.Serial}}

## Retries controls how many times to retry requests to notifiarr.com.
## Sometimes cloudflare returns a 521, and this mitigates those problems.
## Setting this to 0 will take the default of 4. Use 1 to disable retrying.
retries = {{.Retries}}

##################
# Starr Settings #
##################

## The API keys are specific to the app. Get it from Settings -> General.
## Configurations for unused apps are harmless. Set URL and API key for
## apps you have and want to make requests to using Media Bot.
## See the Service Checks section below for information about setting the names.
##
## Examples follow. UNCOMMENT (REMOVE #), AT MINIMUM: [[header]], url, api_key
## Setting any application timeout to "-1s" will disable that application.

{{if .Lidarr}}{{range .Lidarr}}[[lidarr]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"{{if .Username}}
  username = "{{.Username}}"
  password = "{{.Password}}"{{end}}{{if .HTTPUser}}
  http_user = "{{.HTTPUser}}"
  http_pass = "{{.HTTPPass}}"{{end}}
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[lidarr]]
#name     = "" # Set a name to enable checks of your service.
#url      = "http://lidarr:8989/"
#api_key  = ""


{{end}}{{if .Prowlarr}}{{range .Prowlarr}}[[prowlarr]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"{{if .Username}}
  username = "{{.Username}}"
  password = "{{.Password}}"{{end}}{{if .HTTPUser}}
  http_user = "{{.HTTPUser}}"
  http_pass = "{{.HTTPPass}}"{{end}}
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[prowlarr]]
#name     = "" # Set a name to enable checks of your service.
#url      = "http://prowlarr:9696/"
#api_key  = ""


{{end}}{{if .Radarr}}{{range .Radarr}}[[radarr]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"{{if .Username}}
  username = "{{.Username}}"
  password = "{{.Password}}"{{end}}{{if .HTTPUser}}
  http_user = "{{.HTTPUser}}"
  http_pass = "{{.HTTPPass}}"{{end}}
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[radarr]]
#name      = "" # Set a name to enable checks of your service.
#url       = "http://127.0.0.1:7878/"
#api_key   = ""


{{end}}{{if .Readarr}}{{range .Readarr}}[[readarr]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"{{if .Username}}
  username = "{{.Username}}"
  password = "{{.Password}}"{{end}}{{if .HTTPUser}}
  http_user = "{{.HTTPUser}}"
  http_pass = "{{.HTTPPass}}"{{end}}
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[readarr]]
#name      = "" # Set a name to enable checks of your service.
#url       = "http://127.0.0.1:8787/"
#api_key   = ""


{{end}}{{if .Sonarr}}{{range .Sonarr}}[[sonarr]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"{{if .Username}}
  username = "{{.Username}}"
  password = "{{.Password}}"{{end}}{{if .HTTPUser}}
  http_user = "{{.HTTPUser}}"
  http_pass = "{{.HTTPPass}}"{{end}}
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[sonarr]]
#name      = ""  # Set a name to enable checks of your service.
#url       = "http://sonarr:8989/"
#api_key   = ""


{{end -}}

# Download Client Configs (below) are used for dashboard state and service checks.

{{if .Deluge}}{{range .Deluge }}[[deluge]]
  name     = "{{.Name}}"
  url      = "{{.Config.URL}}"
  password = "{{.Password}}"
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[deluge]]
#name     = ""  # Set a name to enable checks of your service.
#url      = "http://deluge:8112/"
#password = ""


{{end}}{{if .Qbit}}{{range .Qbit}}[[qbit]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  user     = "{{.User}}"
  pass     = "{{.Pass}}"
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[qbit]]
#name     = ""  # Set a name to enable checks of your service.
#url      = "http://qbit:8080/"
#user     = ""
#pass     = ""


{{end}}{{if .Rtorrent}}{{range .Rtorrent}}[[rtorrent]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  user     = "{{.User}}"
  pass     = "{{.Pass}}"
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[rtorrent]]
#name     = ""  # Set a name to enable checks of your service.
#url      = "http://rtorrent:5000/"
#user     = ""
#pass     = ""


{{end}}{{if .Transmission}}{{range .Transmission}}[[transmission]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  user     = "{{.User}}"
  pass     = "{{.Pass}}"
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[transmission]]
#name     = ""  # Set a name to enable checks of your service.
#url      = "http://transmission:9091/transmission/rpc"
#user     = ""
#pass     = ""


{{end}}{{if .NZBGet}}{{range .NZBGet}}[[nzbget]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  user     = "{{.User}}"
  pass     = "{{.Pass}}"
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[nzbget]]
#name     = ""  # Set a name to enable checks of your service.
#url      = "http://nzbget:6789/"
#user     = ""
#pass     = ""


{{end}}{{if .SabNZB}}{{range .SabNZB}}[[sabnzbd]]
  name     = "{{.Name}}"
  url      = "{{.URL}}"
  api_key  = "{{.APIKey}}"
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
  timeout  = "{{.Timeout}}"
  {{- if .ValidSSL}}
  valid_ssl = true
  {{- end}}

{{end}}
{{else}}#[[sabnzbd]]
#name     = ""  # Set a name to enable checks of this application.
#url      = "http://sabnzbd:8080/"
#api_key  = ""


{{end -}}

#################
# Plex Settings #
#################

## Find your token: https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/
##
{{if and .Plex (not force)}}[plex]
  url     = "{{.Plex.URL}}"   # Your plex URL
  token   = "{{.Plex.Token}}"   # your plex token; get this from a web inspector
  interval = "{{.Plex.Interval}}" # Service check duration (if name is not empty).
  timeout = "{{.Plex.Timeout}}"  # how long to wait for HTTP responses
  {{- if .Plex.ValidSSL}}
  valid_ssl = true
  {{- end}}
{{- else}}#[plex]
#url     = "http://localhost:32400/" # Your plex URL
#token   = "" # your plex token; get this from a web inspector
{{- end }}

#####################
# Tautulli Settings #
#####################

# Enables email=>username map. Set a name to enable service checks.
# Must uncomment [tautulli], 'api_key' and 'url' at a minimum.
{{if and .Tautulli (not force)}}
[tautulli]
  name     = "{{.Tautulli.Name}}" # only set a name to enable service checks.
  url      = "{{.Tautulli.URL}}" # Your Tautulli URL
  api_key  = "{{.Tautulli.APIKey}}" # your tautulli api key; get this from settings
  timeout  = "{{.Tautulli.Timeout}}" # how long to wait for HTTP responses
  interval = "{{.Tautulli.Interval}}" # how often to send service checks
  {{- if .Tautulli.ValidSSL}}
  valid_ssl = true
  {{- end}}
{{- else}}
#[tautulli]
#  name    = "" # only set a name to enable service checks.
#  url     = "http://localhost:8181/" # Your Tautulli URL
#  api_key = "" # your tautulli api key; get this from settings
{{- end }}

##################
# MySQL Snapshot #
##################

# Enables MySQL process list in snapshot output.
# Adding a name to a server enables TCP service checks.
# Example Grant:
# GRANT PROCESS ON *.* to 'notifiarr'@'localhost'
{{if .Snapshot.MySQL}} {{range .Snapshot.MySQL}}
[[snapshot.mysql]]
  name     = "{{.Name}}"
  host     = "{{.Host}}"
	user     = "{{.User}}"
	pass     = "{{.Pass}}"
  interval = "{{.Interval}}" # Service check duration (if name is not empty).
	timeout  = "{{.Timeout}}"
{{end}}
{{else}}
#[[snapshot.mysql]]
#name = "" # only set a name to enable service checks.
#host = "localhost:3306"
#user = "notifiarr"
#pass = "password"
{{- end}}

###################
# Nvidia Snapshot #
###################

# The app will automatically collect Nvidia data if nvidia-smi is present.
# Use the settings below to disable Nvidia GPU collection, or restrict collection to only specific Bus IDs.
# SMI Path is found automatically if left blank. Set it to path to nvidia-smi (nvidia-smi.exe on Windows).

[snapshot.nvidia]
disabled = {{.Snapshot.Nvidia.Disabled}}
smi_path = '''{{.Snapshot.Nvidia.SMIPath}}'''
bus_ids  = [{{range $s := .Snapshot.Nvidia.BusIDs}}"{{$s}}",{{end}}]

##################
# Service Checks #
##################

## This application performs service checks on configured services at the specified interval.
## The service states are sent to Notifiarr.com. Failed services generate a notification.
## Setting names on Starr apps (above) enables service checks for that app.
## Setting the Interval to "-1s" (Disabled in UI) will disable service checks on that named instance.
## Use the [[service]] directive to add more service checks. Example below.

[services]
  disabled = {{.Services.Disabled}} # Setting this to true disables all service checking routines.
  parallel = {{.Services.Parallel}}     # How many services to check concurrently. 1 should be enough.
  interval = "{{.Services.Interval}}" # How often to send service states to Notifiarr.com. Minimum = 5m.
  log_file = '{{.Services.LogFile}}'    # Service Check logs go to the app log by default. Change that by setting a services.log file here.

## Uncomment the following section to create a service check on a URL or IP:port.
## You may include as many [[service]] sections as you have services to check.
## Do not add Radarr, Sonarr, Readarr, Prowlarr, or Lidarr here! Add a name to enable their checks.
##
## Example with comments follows.
#[[service]]
#  name     = "MyServer"          # name must be unique
#  type     = "http"              # type can be "http" or "tcp"
#  check    = 'http://127.0.0.1/'  # url for 'http', host/IP:port for 'tcp'
#  expect   = "200"               # return code to expect (for http only)
#  timeout  = "10s"               # how long to wait for tcp or http checks.
#  interval = "5m"                # how often to check this service.
{{if not .Service}}
## Another example. Remember to uncomment [[service]] if you use this!
##
#[[service]]
#  name    = "Bazarr"
#  type    = "http"
#  check   = 'http://10.1.1.2:6767/series/'
#  expect  = "200"
#  timeout = "10s"{{else}}
## Configured Service Checks:
##{{range .Service}}
[[service]]
  name     = "{{.Name}}"
  type     = "{{.Type}}"
  check    = '{{.Value}}'
  expect   = "{{.Expect}}"
  timeout  = "{{.Timeout}}"
  interval = "{{.Interval}}"
{{end}}{{end}}


######################
# File & Log Watcher #
######################

## Tail a log file, regex match lines, and send notifications.
## Example:

#[[watch_file]]
#  path  = '/var/log/system.log'
#  skip  = '''error'''
#  regex = '''[Ee]rror'''
#  poll  = false
#  pipe  = false
#  must_exist = false
#  log_match  = true
{{if .WatchFiles}}
## Configured Watch Files:
{{- range $item := .WatchFiles}}{{if $item}}

[[watch_file]]
  path  = '{{$item.Path}}'{{if $item.Skip}}
  skip  = '''{{$item.Skip}}'''{{end}}{{if $item.Regexp}}
  regex = '''{{$item.Regexp}}'''{{end}}{{if $item.Poll}}
  poll  = true{{end}}{{if $item.Pipe}}
  pipe  = true{{end}}{{if $item.MustExist}}
  must_exist = true{{end}}{{if $item.LogMatch}}
  log_match = true{{end}}{{end}}
{{end}}{{end}}


###################
# Custom Commands #
###################

## Run and trigger custom commands.
## Commands may have required arguments thet can be passed in when the command is run.
## These use the format ({regex}) - a regular expression wrapped by curly braces and parens.
## The example below allows a user to run any combination of ls -la on /usr, /home, or /tmp:
## command = "/bin/ls ({-la|-al|-l|-a}) ({/usr|/home|/tmp})"
##
## Full Example (remove the leading # hashes to use it):

#[[command]]
#  name    = 'some-name-for-logs'
#  command = '/var/log/system.log'
#  shell   = false
#  log     = true
#  notify  = true
#  timeout = "10s"
{{if .Commands}}
## Configured Commands:
{{- range $item := .Commands}}{{if $item}}

[[command]]
  name    = '{{$item.Name}}'
  hash    = '{{$item.Hash}}'
  command = '''{{toml $item.Command}}'''
  shell   = {{$item.Shell}}
  log     = {{$item.Log}}
  notify  = {{$item.Notify}}
  timeout = "{{$item.Timeout}}"{{end}}
{{end}}{{end}}
`
