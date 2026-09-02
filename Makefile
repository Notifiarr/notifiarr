# Local development. Releases are GoReleaser Pro — see .github/workflows/README.md.

# Suck in our application information.
IGNORED:=$(shell bash -c "source settings.sh ; env | grep -v BASH_FUNC | sed 's/=/:=/;s/^/export /' > /tmp/.metadata.make")

BUILD_FLAGS=-tags osusergo,netgo
WINDOWS_BUILD_FLAGS=-tags osusergo
GOFLAGS=-trimpath -mod=readonly -modcacherw
CGO_CPPFLAGS=$(CPPFLAGS)
CGO_CFLAGS=$(CFLAGS)
CGO_CXXFLAGS=$(CXXFLAGS)
CGO_LDFLAGS=$(LDFLAGS)

ifeq ($(OUTPUTDIR),)
     OUTPUTDIR=.
endif

# Preserve the passed-in version & iteration (local development testing).
_VERSION:=$(VERSION)
_ITERATION:=$(ITERATION)
include /tmp/.metadata.make

ifneq ($(_VERSION),)
VERSION:=$(_VERSION)
endif

ifneq ($(_ITERATION),)
ITERATION:=$(_ITERATION)
endif

VERSION_LDFLAGS:= -X \"golift.io/version.Branch=$(BRANCH) ($(COMMIT))\" \
	-X \"golift.io/version.BuildDate=$(DATE)\" \
	-X \"golift.io/version.BuildUser=$(shell whoami || echo "unknown")\" \
	-X \"golift.io/version.Revision=$(ITERATION)\" \
	-X \"golift.io/version.Version=$(VERSION)\"

WINDOWS_LDFLAGS:= -H=windowsgui

all: clean generate notifiarr

clean:
	rm -f notifiarr notifiarr.*.{macos,freebsd,linux,exe}{,.gz,.zip} notifiarr.1{,.gz} notifiarr.rb
	rm -f notifiarr{_,-}*.{deb,rpm,txz,zst,sig} v*.tar.gz.sha256 examples/MANUAL .metadata.make rsrc_*.syso
	rm -f cmd/notifiarr/README{,.html} README{,.html} ./notifiarr_manual.html rsrc.syso Notifiarr.*.app.zip
	rm -f notifiarr.service pack.temp.dmg notifiarr.conf.example
	rm -rf package_build_* release Notifiarr.*.app Notifiarr.app dist .docker-build

generate:
	mkdir -p ./frontend/dist
	echo "Fake frontend build." > ./frontend/dist/index.html
	go generate ./frontend/src/api
	go generate ./frontend

man: notifiarr.1.gz
notifiarr.1.gz:
	go run github.com/davidnewhall/md2roff@v0.0.1 --manual notifiarr --version $(VERSION) --date "$(DATE)" examples/MANUAL.md
	gzip -9nc examples/MANUAL > $@
	mv examples/MANUAL.html notifiarr_manual.html

readme: README.html
README.html:
	go run github.com/davidnewhall/md2roff@v0.0.1 --manual notifiarr --version $(VERSION) --date "$(DATE)" README.md

dev: generate main.go
	go build -race $(BUILD_FLAGS) -o $(OUTPUTDIR)/notifiarr -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "
	DN_ENCODE_CONFIG_FILE=false $(OUTPUTDIR)/notifiarr

notifiarr: generate main.go
	go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/notifiarr -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "

test: clean generate lint
	go test -race -covermode=atomic ./...

lint: generate
	codespell -H -L vender,te -S .git,dist,node_modules,fortunes.txt,words.go,swagger*.js,swagger*.map,go.sum,*.json .
	golangci-lint version
	GOOS=linux golangci-lint run
	GOOS=darwin golangci-lint run
	GOOS=freebsd golangci-lint --build-tags nodbus run
	GOOS=windows golangci-lint run

docker: generate
	init/docker/makedocker.sh
