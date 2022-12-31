module github.com/Notifiarr/notifiarr

go 1.19

// home grown goodness.
require (
	golift.io/cache v0.0.2
	golift.io/cnfg v0.2.1
	golift.io/cnfgfile v0.0.0-20221228223038-bba56925665c
	golift.io/datacounter v1.0.4
	golift.io/deluge v0.10.1-0.20221231022948-81f784064173
	golift.io/nzbget v0.1.4
	golift.io/qbit v0.0.0-20221229002737-e0ea34432325
	golift.io/rotatorr v0.0.0-20221229081120-f7a18c5c5533
	golift.io/starr v0.14.1-0.20221230094403-a3c1a181c54a
	golift.io/version v0.0.2
	golift.io/xtractr v0.2.1
)

// menu-ui
require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/gen2brain/beeep v0.0.0-20220909211152-5a9ec94374f6
	github.com/gen2brain/dlgs v0.0.0-20220603100644-40c77870fa8d
	github.com/getlantern/context v0.0.0-20220418194847-3d5e7a086201 // indirect
	github.com/getlantern/errors v1.0.3 // indirect
	github.com/getlantern/golog v0.0.0-20221014032422-49749a7176cf // indirect
	github.com/getlantern/hex v0.0.0-20220104173244-ad7e4b9194dc // indirect
	github.com/getlantern/hidden v0.0.0-20220104173330-f221c5a24770 // indirect
	github.com/getlantern/ops v0.0.0-20220713155959-1315d978fff7 // indirect
	github.com/getlantern/systray v1.2.1
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-toast/toast v0.0.0-20190211030409-01e6764cf0a4 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gonutz/w32 v1.0.0
	github.com/tadvi/systray v0.0.0-20190226123456-11a2b8fa57af // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
)

// snapshot and other stuff.
require (
	github.com/BurntSushi/toml v1.2.1
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-sql-driver/mysql v1.7.0
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/gopherjs/gopherjs v1.17.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/jaypipes/ghw v0.9.0
	github.com/jaypipes/pcidb v1.0.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/lestrrat-go/apache-logformat v0.0.0-20200929122403-cd9b7dc018c7
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/lufia/plan9stats v0.0.0-20220913051719-115f729f3c8c // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20221212215047-62379fc7944b // indirect
	github.com/shirou/gopsutil/v3 v3.22.10
	github.com/spf13/pflag v1.0.6-0.20201009195203-85dd5c8bc61c
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	golang.org/x/mod v0.7.0
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/sys v0.3.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	howett.net/plist v1.0.0 // indirect
)

// zip extraction for database corruption checks.
require (
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/ulikunitz/xz v0.5.11 // indirect
)

// sqlite3 abstraction for database corruption checks.
require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20220927061507-ef77025ab5aa // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/tools v0.4.0 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.40.0 // indirect
	modernc.org/ccgo/v3 v3.16.13 // indirect
	modernc.org/libc v1.22.2 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/opt v0.1.3 // indirect
	modernc.org/sqlite v1.20.1
	modernc.org/strutil v1.1.3 // indirect
	modernc.org/token v1.1.0 // indirect
)

// file watcher stuff.
require (
	github.com/fsnotify/fsnotify v1.6.0
	github.com/gorilla/schema v1.2.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.5.0
	github.com/nxadm/tail v1.4.9-0.20211216163028-4472660a31a6
	golang.org/x/crypto v0.4.0
	golang.org/x/text v0.5.0
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/gorilla/rpc v1.2.0 // indirect
	github.com/hugelgupf/go-shlex v0.0.0-20200702092117-c80c9d0918fa
	github.com/mrobinsn/go-rtorrent v1.8.0
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.8.1
	go.opentelemetry.io/otel v1.11.2 // indirect
	go.opentelemetry.io/otel/trace v1.11.2 // indirect
)

// api docs.
require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.7 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/swaggo/swag v1.8.9
	github.com/urfave/cli/v2 v2.23.7 // indirect
)

// ping service check
require (
	github.com/go-ping/ping v1.1.0
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/sync v0.1.0 // indirect
)

// tools.
require (
	github.com/golang/mock v1.6.0
	github.com/kevinburke/go-bindata v3.24.0+incompatible
)

// xtractr stuff...
require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/bodgit/plumbing v1.3.0 // indirect
	github.com/bodgit/sevenzip v1.4.0 // indirect
	github.com/bodgit/windows v1.0.1 // indirect
	github.com/connesc/cipherio v0.2.1 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/klauspost/compress v1.15.13 // indirect
	github.com/pierrec/lz4/v4 v4.1.17 // indirect
	go4.org v0.0.0-20201209231011-d4a079459e60 // indirect
)

require github.com/kdomanski/iso9660 v0.3.3 // indirect
