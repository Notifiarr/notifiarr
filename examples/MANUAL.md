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

`notifiarr [-c <file>] [-w <file>] [--ps] [--curl <url> [--header <header>]] [-h] [-v]`

    -c, --config <config file>
        Provide a configuration file (instead of the default).
        Can also be passed as environment variable: DN_CONFIG_FILE

    -e, --extraconfig <config file>[,<config file>[,<config file> ...]]
        You may load in multiple config files. Useful for storing
        passwords and api keys in a different location.

    -p, --prefix
        The default environment variable configuration prefix is `DN`.
        Use this tunable to change it.

    --ps
        This flags makes the application output the system process list and exit.
        Useful for debugging and testing 'process' service checks.

    -w, --write <file>
        This flag allows you to read in a config file and re-write it to another file.
        Use - as a shortcut argument to over write the file provided by --config.
        This will not touch config files added with --extraconfig, but it will
        write their content/settings into the new config (now combined) file.

    --curl <url>
        This flags allows you to make the application GET a URL and print
        the response. Very simple and similar to curl.

    --header <HTTP Request Header>
        This flag only works with --curl. Use this to pass a request header when
        making an http request. Example: --header "X-Plex-Token: absdefg123jklm"
        May be provided more than once to add multiple headers.

    -a, --assets <folder/path>
        Use this flag to pass in a custom HTTP assets folder. This folder must
        contain a files/ folder and a templates/ folder. Files are served as
        static assets from the /files path and templates contains customized
        HTML templates for the GUI web routes. This is an advanced feature.

    --apthook
        This should only be used on Linux for dpkg pre-install-pkg hooks.
        This reads from stdin from dpkg, and sends it off to notifiarr.com.
        Will not work on any OS except Linux; only designed for Debian OSes.

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
