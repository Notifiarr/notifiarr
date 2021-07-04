notifiarr(1) -- Unified Client for Notifiarr.com
===

SYNOPSIS
---

`notifiarr -c /etc/notifiarr/notifiarr.conf`

This service runs a web server that allows notifiarr.com's Media Bot to
communicate with Radarr, Lidarr, Readarr and Sonarr. This provides the ability
to add new content to your libraries from within Discord. This client also sends
system snapshot and Plex session data to Notifiarr for Discord notifications.

OPTIONS
---

`notifiarr [-c <file>] [--write <file>] [--snaps] [--ps] [--cfsync] [--curl <url>] [-h] [-v]`

    -c, --config <config file>
        Provide a configuration file (instead of the default).
        Can also be passed as environment variable: DN_CONFIG_FILE

    -p, --prefix
        The default environment variable configuration prefix is `DN`.
        Use this tunable to change it.

    --snaps
        This flag makes the application output a system snapshot and exit.
        Useful for debugging and testing.

    --ps
        This flags makes the application output the system process list and exit.
        Useful for debugging and testing 'process' service checks.

    --cfsync
        This flag makes the application send a Custom Format request to
        Notifiarr.com. Then it exits.

    --write <file>
        This flag allows you to read in a config file and re-write it to another file.
        Use - as a shortcut argument to over write the file provided by --config.

    --curl <url>
        This flags allows you to make the application GET a URL and print
        the response. Very simple and similar to curl.

    -v, --version
        Display version and exit.

    -h, --help
        Display usage and exit.

CONFIGURATION
---

The default configuration file location changes depending on operating system.
See the provided example configuration file for parameter information.

INPUT
---

On Windows and with the App version on macOS you can work with the app using the
system tray menu. On Unix OSes you can send a USR1 signal to re-write the config
file using a built-in template. Backup your existing config file first.

AUTHOR
---
*   David Newhall II - 12/2020

LOCATION
---
* Notifiarr: [https://github.com/Notifiarr/notifiarr](https://github.com/Notifiarr/notifiarr)
