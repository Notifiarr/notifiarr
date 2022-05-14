module github.com/Notifiarr/notifiarr

go 1.17

// home grown goodness.
require (
	golift.io/cnfg v0.2.0
	golift.io/cnfgfile v0.0.0-20220228094812-5864a2124d02
	golift.io/deluge v0.9.4-0.20220317021225-51ab09943349
	golift.io/qbit v0.0.0-20220317021235-151a7d2d428e
	golift.io/rotatorr v0.0.0-20220126065426-d6b5acaac41c
	golift.io/starr v0.14.1-0.20220227102118-e029bb583eec
	golift.io/version v0.0.2
	golift.io/xtractr v0.0.11
)

// menu-ui
require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/gen2brain/beeep v0.0.0-20210529141713-5586760f0cc1
	github.com/gen2brain/dlgs v0.0.0-20211108104213-bade24837f0b
	github.com/getlantern/context v0.0.0-20190109183933-c447772a6520 // indirect
	github.com/getlantern/errors v1.0.1 // indirect
	github.com/getlantern/golog v0.0.0-20211223150227-d4d95a44d873 // indirect
	github.com/getlantern/hex v0.0.0-20220104173244-ad7e4b9194dc // indirect
	github.com/getlantern/hidden v0.0.0-20220104173330-f221c5a24770 // indirect
	github.com/getlantern/ops v0.0.0-20200403153110-8476b16edcd6 // indirect
	github.com/getlantern/systray v1.2.1
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-toast/toast v0.0.0-20190211030409-01e6764cf0a4 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/gonutz/w32 v1.0.0
	github.com/iamacarpet/go-win64api v0.0.0-20220314100901-d3a958911279
	github.com/tadvi/systray v0.0.0-20190226123456-11a2b8fa57af // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	gopkg.in/toast.v1 v1.0.0-20180812000517-0a84660828b2 // indirect
)

// snapshot and other stuff.
require (
	github.com/BurntSushi/toml v1.0.0
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-sql-driver/mysql v1.6.0
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/google/cabbie v1.0.3 // indirect
	github.com/google/glazier v0.0.0-20220309125052-ca3c88631db6 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20220221023154-0b2280d3ff96 // indirect
	github.com/gopherjs/gopherwasm v1.1.0 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/jaypipes/ghw v0.8.0
	github.com/jaypipes/pcidb v0.6.0 // indirect
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/lestrrat-go/apache-logformat v0.0.0-20200929122403-cd9b7dc018c7
	github.com/lestrrat-go/strftime v1.0.5 // indirect
	github.com/lufia/plan9stats v0.0.0-20220305071607-d0b38dbe16db // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/nu7hatch/gouuid v0.0.0-20131221200532-179d4d0c4d8d // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20220216144756-c35f1ee13d7c // indirect
	github.com/scjalliance/comshim v0.0.0-20190308082608-cf06d2532c4e // indirect
	github.com/shirou/gopsutil/v3 v3.22.2
	github.com/spf13/pflag v1.0.6-0.20201009195203-85dd5c8bc61c
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220315194320-039c03cc5b86
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	howett.net/plist v1.0.0 // indirect
)

// zip extraction for databsae corruption checks.
require (
	github.com/nwaples/rardecode v1.1.3 // indirect
	github.com/saracen/go7z v0.0.0-20191010121135-9c09b6bd7fda // indirect
	github.com/saracen/go7z-fixtures v0.0.0-20190623165746-aa6b8fba1d2f // indirect
	github.com/saracen/solidblock v0.0.0-20190426153529-45df20abab6f // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
)

// sqlite3 abstraction for databsae corruption checks.
require (
	github.com/google/uuid v1.3.0 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20200410134404-eec4a21b6bb0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/tools v0.1.10 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	lukechampine.com/uint128 v1.2.0 // indirect
	modernc.org/cc/v3 v3.35.24 // indirect
	modernc.org/ccgo/v3 v3.15.17 // indirect
	modernc.org/libc v1.14.11 // indirect
	modernc.org/mathutil v1.4.1 // indirect
	modernc.org/memory v1.0.6 // indirect
	modernc.org/opt v0.1.1 // indirect
	modernc.org/sqlite v1.15.2
	modernc.org/strutil v1.1.1 // indirect
	modernc.org/token v1.0.0 // indirect
)

require (
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/fsnotify/fsnotify v1.5.1
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/schema v1.2.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.5.0
	github.com/nxadm/tail v1.4.8
	golang.org/x/crypto v0.0.0-20220315160706-3147a52a75dd
	golift.io/datacounter v1.0.3
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)

