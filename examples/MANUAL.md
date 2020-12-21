discordnotifier-client(1) -- Unified Client for DiscordNotifier.com
===

SYNOPSIS
---

`discordnotifier-client -c /etc/discordnotifier-client/dnconfig.conf`

This service runs a web server that allows discordnotifier.com's Media Bot to
communicate with Radarr, Lidarr, Readarr and Sonarr. This provides the ability
to add new content to your libraries from within Discord.

OPTIONS
---

`discordnotifier-client [-c <config file>] [-h] [-v]`

    -c, --config <config file>
        Provide a configuration file (instead of the default).

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
*   DiscordNotifier Client: [https://github.com/Go-Lift-TV/discordnotifier-client](https://github.com/Go-Lift-TV/discordnotifier-client)
