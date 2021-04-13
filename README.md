<img src="https://docs.notifiarr.com/img/repo-logo.png">

This is the unified client for [Notifiarr.com](https://notifiarr.com).
The client enables content requests from Media Bot in your Discord Server.
It also provides reports for Plex usage and system health.

## Installation

### Linux

[![packagecloud](https://docs.golift.io/integrations/packagecloud.png "PackageCloud.io")](http://packagecloud.io)

Linux repository hosting provided by [Packagecloud.io](https://packagecloud.io)!

After install, edit the config and start the service:

```
sudo nano /etc/notifiarr/notifiarr.conf
sudo service systemctl restart notifiarr
```

#### Debian Variants

- Ubuntu, etc.

Install the repo like this. All variants use the same `debian/focal` repo. The app works on all Linuxes.

```shell
curl -L https://packagecloud.io/golift/pkgs/gpgkey | sudo apt-key add -
echo "deb https://packagecloud.io/golift/pkgs/ubuntu focal main" | sudo tee /etc/apt/sources.list.d/golift.conf
sudo apt update
sudo apt install notifiarr
```

#### RedHat Variants

- CentOS, Fedora, SUSE, etc.

Install the repo like this. All variants use the same `el/6` repo. The app works on all Linuxes.

```shell
sudo tee golift.repo <<-EOF
[golift]
name=golift
baseurl=https://packagecloud.io/golift/pkgs/el/6/\$basearch
repo_gpgcheck=1
gpgcheck=1
enabled=1
gpgkey=https://packagecloud.io/golift/pkgs/gpgkey
       https://packagecloud.io/golift/pkgs/gpgkey/golift-pkgs-7F7791485BF8996D.pub.gpg
sslverify=1
sslcacert=/etc/pki/tls/certs/ca-bundle.crt
metadata_expire=300
EOF

sudo yum -q makecache -y --disablerepo='*' --enablerepo='golift'
sudo yum install notifiarr
```

### FreeBSD

-   Download a package from the [Releases](https://github.com/Notifiarr/notifiarr/releases) page.
-   Install it, edit config, start it.

Example of the above in shell form:

```shell
wget -qO- https://raw.githubusercontent.com/Notifiarr/notifiarr/main/scripts/install.sh | sudo bash

vi /usr/local/etc/notifiarr/notifiarr.conf
service notifiarr start
```

On Linux and FreeBSD, Notifiarr runs as `user:group` `notifiarr:notifiarr`.

### macOS Install

#### Homebrew

-   Edit config file at `/usr/local/etc/notifiarr/notifiarr.conf`
-   Start it.
-   Like this:

```shell
brew install golift/mugs/notifiarr
vi /usr/local/etc/notifiarr/notifiarr.conf
brew services start notifiarr
```

#### macOS App

-   You can use the unsigned app on the Releases page.
-   You must right click the app and select `Open` so macOS allows it.
-   When you open it for the first time it will create a config file and log file:
    -   `~/.notifiarr/notifiarr.conf`
    -   `~/.notifiarr/notifiarr.log`
-   Edit the config file and reload or restart the app.

### Windows

-   Extract a `.exe.zip` file from [the Releases page](https://github.com/Notifiarr/notifiarr/releases).
-   Run the `notifiarr.amd64.exe` binary. This starts the app in the system tray.
-   When you open it for the first time it will create a config file and log file:
    -   `C:\ProgramData\notifiarr\notifiarr.conf`
    -   `<your home folder>\.notifiarr\notifiarr.log`
-   Edit the new config file suit your environment then reload or restart the app.

### Docker

This project builds automatically in [Docker Cloud](https://hub.docker.com/r/golift/notifiarr)
and creates [ready-to-use multi-architecture images](https://hub.docker.com/r/golift/notifiarr/tags).
The `latest` tag is always a tagged release in GitHub. The `main` tag corresponds
to the `main` branch in GitHub and may be broken.

#### Docker Config File

-   Copy the [example config file](https://github.com/Notifiarr/notifiarr/blob/main/examples/notifiarr.conf.example) from this repo.
-   Then grab the image from docker hub and run it using an overlay for the config file.
-   You must set `privileged` to use `smartctl` (`monitor_drives`) and/or `MegaCli` (`monitor_raid`).
-   Map the `/var/run/utmp` volume if you want to count users.
-   Mount any volumes you want to report storage space for. Where does not matter, "where" is the "name".

```shell
docker pull golift/notifiarr
docker run -d \
-v /your/config/notifiarr.conf:/config/notifiarr.conf \
-v /var/run/utmp:/var/run/utmp \
golift/notifiarr
docker logs <container id from docker run>
```

#### Docker Environment Variables

See below for more information about which environment variables are available.
You must set `--privileged` when `monitor_drives=true`.

```shell
docker pull golift/notifiarr
docker run -d --privileged \
  -v /var/run/utmp:/var/run/utmp \
  -e "DN_API_KEY=abcdef-12345-bcfead-43312-bbbaaa-123" \
  -e "DN_SONARR_0_URL=http://localhost:8989" \
  -e "DN_SONARR_0_API_KEY=kjsdkasjdaksdj" \
  -e "DN_SNAPSHOT_MONITOR_DRIVES=true" \
  golift/notifiarr
docker logs <container id from docker run>
```

## Configuration Information

-   Instead of, or in addition to a config file, you may configure a docker
    container with environment variables.
-   Any variable not provided takes the default.
-   Must provide an API key from notifiarr.com.
    -   **The Notifiarr application uses the API key for bi-directional authorization.**
-   Must provide URL and API key for Sonarr or Radarr or Readarr or any combination.
-   You may provide multiple sonarr, radarr or readarr instances using
    `DN_SONARR_1_URL`, `DN_SONARR_2_URL`, etc.

|Config Name|Variable Name|Default / Note|
|---|---|---|
api_key|`DN_API_KEY`|**Required** / API Key from Notifiarr.com|
bind_addr|`DN_BIND_ADDR`|`0.0.0.0:5454` / The IP and port to listen on|
quiet|`DN_QUIET`|`false` / Turns off output. Set a log_file if this is true|
urlbase|`DN_URLBASE`|default: `/` Change the web root with this setting|
upstreams|`DN_UPSTREAMS_0`|List of upstream networks that can set X-Forwarded-For|
ssl_key_file|`DN_SSL_KEY_FILE`|Providing SSL files turns on the SSL listener|
ssl_cert_file|`DN_SSL_CERT_FILE`|Providing SSL files turns on the SSL listener|
log_file|`DN_LOG_FILE`|None by default. Optionally provide a file path to save app logs|
http_log|`DN_HTTP_LOG`|None by default. Provide a file path to save HTTP request logs|
log_file_mb|`DN_LOG_FILE_MB`|`100` / Max size of log files in megabytes|
log_files|`DN_LOG_FILES`|`10` / Log files to keep after rotating. `0` disables rotation|
timeout|`DN_TIMEOUT`|`60s` / Global API Timeouts (all apps default)|

#### Sonarr

|Config Name|Variable Name|Note|
|---|---|---|
sonarr.name|`DN_SONARR_0_NAME`|No Default. Setting a name enabled service checks.|
sonarr.url|`DN_SONARR_0_URL`|No Default. Something like: `http://localhost:8989`|
sonarr.api_key|`DN_SONARR_0_API_KEY`|No Default. Provide URL and API key if you use Sonarr|

#### Radarr

|Config Name|Variable Name|Note|
|---|---|---|
radarr.name|`DN_RADARR_0_NAME`|No Default. Setting a name enabled service checks.|
radarr.url|`DN_RADARR_0_URL`|No Default. Something like: `http://localhost:7878`|
radarr.api_key|`DN_RADARR_0_API_KEY`|No Default. Provide URL and API key if you use Radarr|

#### Readarr

|Config Name|Variable Name|Note|
|---|---|---|
readarr.name|`DN_READARR_0_NAME`|No Default. Setting a name enabled service checks.|
readarr.url|`DN_READARR_0_URL`|No Default. Something like: `http://localhost:8787`|
readarr.api_key|`DN_READARR_0_API_KEY`|No Default. Provide URL and API key if you use Readarr|

#### Lidarr

|Config Name|Variable Name|Note|
|---|---|---|
lidarr.name|`DN_LIDARR_0_NAME`|No Default. Setting a name enabled service checks.|
lidarr.url|`DN_LIDARR_0_URL`|No Default. Something like: `http://lidarr:8686`|
lidarr.api_key|`DN_LIDARR_0_API_KEY`|No Default. Provide URL and API key if you use Readarr|

#### Plex

This application can also send Plex sessions to Notfiarr so you can receive
notifications when users interact with your server. This has three different features:

- Notify all sessions on a longer interval (30+ minutes).
- Notify on session nearing completion (percent complete).
- Notify on session change (Plex Webhook) ie. pause/resume.

You [must provide Plex Token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/)
for this to work. Setting `movies_percent_complete` or `series_percent_complete` to a number above 0 will cause this
application to poll Plex once per minute looking for sessions nearing completion. If Plex goes down
this will cause a lot of log spam. You may also need to add a webhook to Plex so it sends notices to this application.

- In Plex Media Server, add this URL to webhooks:
  - `http://localhost:5454/plex?token=plex-token-here`
- Replace `localhost` with the IP or host of the notifiarr application.
- Replace `plex-token-here` with your plex token.
- **The Notifiarr application uses the Plex token to authorize incoming webhooks.**

|Config Name|Variable Name|Note|
|---|---|---|
plex.url|`DN_PLEX_URL`|`http://localhost:32400` / local URL to your plex server|
plex.token|`DN_PLEX_TOKEN`|Required. [Must provide Plex Token](https://support.plex.tv/articles/204059436-finding-an-authentication-token-x-plex-token/) for this to work.|
plex.interval|`DN_PLEX_INTERVAL`|`30m`, How often to notify on all session data (cron)|
plex.cooldown|`DN_PLEX_COOLDOWN`|`10s`, Maximum rate of notifications is 1 every cooldown interval|
plex.account_map|`DN_PLEX_ACCOUNT_MAP`|map an email to a name, ex: `"som@ema.il,Name|some@ther.mail,name"`|
plex.server|`DN_PLEX_SERVER`|Optional name of this server; discovered automatically otherwise|
plex.movies_percent_complete|`DN_PLEX_MOVIES_PERCENT_COMPLETE`|Send complete notice when a movie reaches this percent.|
plex.series_percent_complete|`DN_PLEX_SERIES_PERCENT_COMPLETE`|Send complete notice when a show reaches this percent.|

#### System Snapshot

This application can also take a snapshot of your system at an interval and send
you a notification. Snapshot means system health like cpu, memory, disk, raid, users, etc.

If you monitor drive health you must have smartmontools (`smartctl`) installed.
If you use smartctl on Linux, you must enable sudo. Add this sudoers entry to
`/etc/sudoers` and fix the path to `smartctl` if yours differs. If you monitor
raid and use MegaCli (LSI card), add the appropriate sudoers entry for that too.

```
notifiarr ALL=(root) NOPASSWD:/usr/sbin/smartctl *
notifiarr ALL=(root) NOPASSWD:/usr/sbin/MegaCli64 -LDInfo -Lall -aALL
```

###### Snapshot Packages

  - **Windows**:  `smartmontools` - get it here https://sourceforge.net/projects/smartmontools/
  - **Linux**:    Debian/Ubuntu: `apt install smartmontools`, RedHat/CentOS: `yum install smartmontools`
  - **Synology**: `opkg install smartmontools`, but first get Entware:
    - Entware (synology):  https://github.com/Entware/Entware-ng/wiki/Install-on-Synology-NAS
    - Entware Package List:  https://github.com/Entware/Entware-ng/wiki/Install-on-Synology-NAS

###### Snapshot Configuration

|Config Name|Variable Name|Note|
|---|---|---|
snapshot.interval|`DN_SNAPSHOT_INTERVAL`|`30m`, How often to send a snapshot (cron)|
snapshot.timeout|`DN_SNAPSHOT_TIMEOUT`|`10s`, How long to wait for a reply from Notifiarr.com|
snapshot.monitor_raid|`DN_SNAPSHOT_MONITOR_RAID`|Set `true` to report `mdadm` and `megacli`|
snapshot.monitor_drives|`DN_SNAPSHOT_MONITOR_DRIVES`|Set `true` to report SMART on drives|
snapshot.monitor_space|`DN_SNAPSHOT_MONITOR_SPACE`|Set `true` to report drive volume usage|
snapshot.monitor_uptime|`DN_SNAPSHOT_MONITOR_UPTIME`|Set `true` to report Local Host Information|
snapshot.monitor_cpuMemory|`DN_SNAPSHOT_MONITOR_CPUMEMORY`|Set `true` to report CPU and Memory usage|
snapshot.monitor_cpuTemp|`DN_SNAPSHOT_MONITOR_CPUTEMP`|Set `true` to report CPU temperatures|
snapshot.zfs_pools|`DN_SNAPSHOT_ZFS_POOL_0`|Provide a list of zfs pools to monitor, ie. `data`|
snapshot.use_sudo|`DN_SNAPSHOT_USE_SUDO`|Set `true` if `monitor_drives=true` or you use `megacli` on Linux|

- _Notes: Not all systems can report CPU temperatures._

#### Service Checks

The Notifiarr client can also check URLs for health. If you set names on your
Starr apps they will be automatically checked and reports sent to Notifiarr.

|Config Name|Variable Name|Note|
|---|---|---|
services.interval|`DN_SERVICES_INTERVAL`|`10m`, How often to check service health; minimum: `5m`|
services.parallel|`DN_SERVICES_PARALLE`|`1`, How many services can be checked at once; 1 is plenty|

You can also create ad-hoc service checks for things like Bazarr.

|Config Name|Variable Name|Note|
|---|---|---|
service.name|`DN_SERVICE_0_NAME`|Services must have a unique name|
service.type|`DN_SERVICE_0_TYPE`|Type must be one of `http`, `tcp`|
service.check|`DN_SERVICE_0_CHECK`|The `URL`, or `host/ip:port` to check|
service.expect|`DN_SERVICE_0_EXPECT`|`200`, For HTTP, the return code to expect|
service.timeout|`DN_SERVICE_0_TIMEOUT`|`15s`, How long to wait for service response|

## Reverse Proxy

You'll need to expose this application to the Internet, so Notifiarr.com
can make connections to it. While you can certainly poke a hole your firewall
and send the traffic directly to this app, it is recommended that you put it
behind a reverse proxy. It's pretty easy.

You'll want to tune the `upstreams` and `urlbase` settings for your environment.
If your reverse proxy IP is `192.168.3.45` then set `upstreams = ["192.168.3.45/32"]`.
The `urlbase` can be left at `/`, but change it if you serve this app from a
subfolder. We'll assume you want to serve the app from `/notifiarr/` and
it's running on `192.168.3.33` - here's a sample nginx config to do that:

```
location /notifiarr/ {
  proxy_set_header X-Forwarded-For $remote_addr;
  proxy_pass http://192.168.3.33:5454$request_uri;
}
```

Make sure the `location` path matches the `urlbase` and ends with a `/`.
That's all there is to it.

## Troubleshooting

- Find help on [GoLift Discord](https://golift.io/discord).
- And/or on [Notifiarr Discord](https://discord.gg/AURf8Yz).

Log files:

You can set a log file in the config. You should do that. Otherwise, find your logs here:

-   Linux: `/var/log/messages` or `/var/log/syslog` (w/ default syslog)
-   FreeBSD: `/var/log/syslog` (w/ default syslog)
-   Homebrew: `/usr/local/var/log/notifiarr.log`
-   macOS: `~/.notifiarr/notifiarr.log`
-   Windows: `<home folder>/.notifiarr/notifiarr.log`

If transfers are in a Warning or Error state they will not be extracted.

Still having problems?
[Let us know!](https://github.com/Notifiarr/notifiarr/issues/new)

## Contributing

Yes, please.

## License

[MIT](https://github.com/Notifiarr/notifiarr/blob/main/LICENSE) - Copyright (c) 2020-2021 Go Lift
