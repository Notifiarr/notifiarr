notifiarr(1) -- Unified Client for Notifiarr.com
===

SYNOPSIS
---

`notifiarr -c /etc/notifiarr/notifiarr.conf`

This service runs a web server that allows notifiarr.com's Media Bot to
communicate with Radarr, Lidarr, Readarr and Sonarr. This provides the ability
to add new content to your libraries from within Discord. This client also sends
system snapshot and Plex session data to Notifiarr for Discord notifictaions.

OPTIONS
---

`notifiarr [-c <config file>] [-h] [-v]`

    -c, --config <config file>
        Provide a configuration file (instead of the default).
        Can also be passed as environment variable: DN_CONFIG_FILE

    -p, --prefix
        The default environment variable configuration prefix is `DN`.
        Use this tunable to change it.

    -v, --version
        Display version and exit.

    -h, --help
        Display usage and exit.

CONFIGURATION
---

The default configuration file location changes depending on operating system.
See the provided example configuration file for parameter information.

AUTHOR
---
*   David Newhall II - 12/2020

LOCATION
---
* Notifiarr: [https://github.com/Go-Lift-TV/notifiarr](https://github.com/Go-Lift-TV/notifiarr)
