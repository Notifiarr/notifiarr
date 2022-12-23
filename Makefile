# This Makefile is written as generic as possible.
# Setting the variables in settings.sh and creating the paths in the repo makes this work.
# See more: https://github.com/golift/application-builder

# Suck in our application information.
IGNORED:=$(shell bash -c "source settings.sh ; env | grep -v BASH_FUNC | sed 's/=/:=/;s/^/export /' > /tmp/.metadata.make")

EXTRA_FPM_FLAGS=--conflicts=discordnotifier-client>0.0.1 --provides=notifiarr --provides=discordnotifier-client
BUILD_FLAGS=-tags osusergo,netgo
GOFLAGS=-trimpath -mod=readonly -modcacherw
CGO_CPPFLAGS=$(CPPFLAGS)
CGO_CFLAGS=$(CFLAGS)
CGO_CXXFLAGS=$(CXXFLAGS)
CGO_LDFLAGS=$(LDFLAGS)

ifeq ($(OUTPUTDIR),)
     OUTPUTDIR=.
endif

# Preserve the passed-in version & iteration (homebrew).
_VERSION:=$(VERSION)
_ITERATION:=$(ITERATION)
include /tmp/.metadata.make

# Travis CI passes the version in. Local builds get it from the current git tag.
ifneq ($(_VERSION),)
VERSION:=$(_VERSION)
ITERATION:=$(_ITERATION)
endif

# rpm is weird and changes - to _ in versions.
RPMVERSION:=$(shell echo $(VERSION) | tr -- - _)

PACKAGE_SCRIPTS=--before-install init/systemd/before-install.sh \
--after-install init/systemd/after-install.sh \
--before-remove init/systemd/before-remove.sh

define PACKAGE_ARGS
$(PACKAGE_SCRIPTS) \
--name notifiarr \
--deb-no-default-config-files \
--rpm-os linux \
--deb-user notifiarr \
--rpm-user notifiarr \
--pacman-user notifiarr \
--iteration $(ITERATION) \
--license $(LICENSE) \
--url $(SOURCE_URL) \
--maintainer "$(MAINT)" \
--vendor "$(VENDOR)" \
--description "$(DESC)" \
--config-files "/etc/notifiarr/notifiarr.conf" \
--freebsd-origin "$(SOURCE_URL)"
endef

VERSION_LDFLAGS:= -X \"golift.io/version.Branch=$(BRANCH) ($(COMMIT))\" \
	-X \"golift.io/version.BuildDate=$(DATE)\" \
	-X \"golift.io/version.BuildUser=$(shell whoami || echo "unknown")\" \
	-X \"golift.io/version.Revision=$(ITERATION)\" \
	-X \"golift.io/version.Version=$(VERSION)\"

# Makefile targets follow.

all: clean build

####################
##### Releases #####
####################

# Prepare a release. Called in Travis CI.
release: clean generate linux_packages freebsd_packages windows
	# Prepareing a release!
	mkdir -p $@
	mv notifiarr.*.linux notifiarr.*.freebsd $@/
	gzip -9r $@/
	for i in notifiarr*.exe ; do zip -9qj $@/$$i.zip $$i examples/*.example *.html; rm -f $$i;done
	mv *.rpm *.deb *.txz $@/
	# Generating File Hashes
	openssl dgst -r -sha256 $@/* | sed 's#release/##' | tee $@/checksums.sha256.txt

# requires a mac.
signdmg: Notifiarr.app
	bash init/macos/makedmg.sh

# Delete all build assets.
clean:
	rm -f notifiarr notifiarr.*.{macos,freebsd,linux,exe,upx}{,.gz,.zip} notifiarr.1{,.gz} notifiarr.rb
	rm -f notifiarr{_,-}*.{deb,rpm,txz} v*.tar.gz.sha256 examples/MANUAL .metadata.make rsrc_*.syso
	rm -f cmd/notifiarr/README{,.html} README{,.html} ./notifiarr_manual.html rsrc.syso Notifiarr.*.app.zip
	rm -f notifiarr.service pkg/bindata/bindata.go pack.temp.dmg
	rm -rf package_build_* release Notifiarr.*.app Notifiarr.app
	rm -f pkg/bindata/docs/api_docs.go

####################
##### Sidecars #####
####################

# Build a man page from a markdown file using md2roff.
# This also turns the repo readme into an html file.
# md2roff is needed to build the man file and html pages from the READMEs.
man: notifiarr.1.gz
notifiarr.1.gz:
	# Building man page. Build dependency first: md2roff
	$(shell go env GOPATH)/bin/md2roff --manual notifiarr --version $(VERSION) --date "$(DATE)" examples/MANUAL.md
	gzip -9nc examples/MANUAL > $@
	mv examples/MANUAL.html notifiarr_manual.html

readme: README.html
README.html: 
	# This turns README.md into README.html
	$(shell go env GOPATH)/bin/md2roff --manual notifiarr --version $(VERSION) --date "$(DATE)" README.md

rsrc: rsrc.syso
rsrc.syso: init/windows/application.ico init/windows/manifest.xml
	$(shell go env GOPATH)/bin/rsrc -arch amd64 -ico init/windows/application.ico -manifest init/windows/manifest.xml

####################
##### Binaries #####
####################

build: notifiarr
notifiarr: generate main.go
	go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/notifiarr -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "

linux: notifiarr.amd64.linux
notifiarr.amd64.linux:  main.go
	# Building linux 64-bit x86 binary.
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/$@ -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "

linux386: notifiarr.386.linux
notifiarr.386.linux:  main.go
	# Building linux 32-bit x86 binary.
	GOOS=linux GOARCH=386 go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/$@ -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "

arm: arm64 armhf

arm64: notifiarr.arm64.linux
notifiarr.arm64.linux:  main.go
	# Building linux 64-bit ARM binary.
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/$@ -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "
	# https://github.com/upx/upx/issues/351#issuecomment-599116973

armhf: notifiarr.arm.linux
notifiarr.arm.linux:  main.go
	# Building linux 32-bit ARM binary.
	GOOS=linux GOARCH=arm GOARM=6 go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/$@ -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "

macos: notifiarr.universal.macos
notifiarr.universal.macos: notifiarr.amd64.macos notifiarr.arm64.macos
	# Building darwin 64-bit universal binary.
	lipo -create -output $@ notifiarr.amd64.macos notifiarr.arm64.macos
notifiarr.amd64.macos:  main.go
	# Building darwin 64-bit x86 binary.
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CGO_LDFLAGS=-mmacosx-version-min=10.8 CGO_CFLAGS=-mmacosx-version-min=10.8 go build -o $@ -ldflags "-v -w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "
notifiarr.arm64.macos: generate main.go
	# Building darwin 64-bit arm binary.
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 CGO_LDFLAGS=-mmacosx-version-min=10.8 CGO_CFLAGS=-mmacosx-version-min=10.8 go build -o $@ -ldflags "-v -w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "


freebsd: notifiarr.amd64.freebsd
notifiarr.amd64.freebsd: generate main.go
	GOOS=freebsd GOARCH=amd64 go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/$@ -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "

freebsd386: notifiarr.i386.freebsd
notifiarr.i386.freebsd: generate main.go
	GOOS=freebsd GOARCH=386 go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/$@ -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "

freebsdarm: notifiarr.armhf.freebsd
notifiarr.armhf.freebsd: generate main.go
	GOOS=freebsd GOARCH=arm go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/$@ -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) "

exe: notifiarr.amd64.exe
windows: notifiarr.amd64.exe
notifiarr.amd64.exe: generate rsrc.syso main.go
	# Building windows 64-bit x86 binary.
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(OUTPUTDIR)/$@ -ldflags "-w -s $(VERSION_LDFLAGS) $(EXTRA_LDFLAGS) $(WINDOWS_LDFLAGS)"

####################
##### Packages #####
####################

linux_packages: rpm deb rpm386 deb386 debarm rpmarm debarmhf rpmarmhf

freebsd_packages: freebsd_pkg freebsd386_pkg freebsdarm_pkg

macapp: Notifiarr.app
Notifiarr.app: notifiarr.universal.macos
	cp -rp init/macos/Notifiarr.app Notifiarr.app
	mkdir -p Notifiarr.app/Contents/MacOS
	cp notifiarr.universal.macos Notifiarr.app/Contents/MacOS/Notifiarr
	sed -i '' -e "s/{{VERSION}}/$(VERSION)/g" Notifiarr.app/Contents/Info.plist

rpm: notifiarr-$(RPMVERSION)-$(ITERATION).x86_64.rpm
notifiarr-$(RPMVERSION)-$(ITERATION).x86_64.rpm: package_build_linux_rpm check_fpm
	@echo "Building 'rpm' package for notifiarr version '$(RPMVERSION)-$(ITERATION)'."
	fpm -s dir -t rpm $(PACKAGE_ARGS) -a x86_64 -v $(RPMVERSION) -C $< $(EXTRA_FPM_FLAGS)
	[ "$(SIGNING_KEY)" = "" ] || rpmsign --key-id=$(SIGNING_KEY) --resign notifiarr-$(RPMVERSION)-$(ITERATION).x86_64.rpm

deb: notifiarr_$(VERSION)-$(ITERATION)_amd64.deb
notifiarr_$(VERSION)-$(ITERATION)_amd64.deb: package_build_linux_deb check_fpm
	@echo "Building 'deb' package for notifiarr version '$(VERSION)-$(ITERATION)'."
	fpm -s dir -t deb $(PACKAGE_ARGS) -a amd64 -v $(VERSION) -C $< $(EXTRA_FPM_FLAGS)
	[ "$(SIGNING_KEY)" = "" ] || debsigs --default-key="$(SIGNING_KEY)" --sign=origin notifiarr_$(VERSION)-$(ITERATION)_amd64.deb

rpm386: notifiarr-$(RPMVERSION)-$(ITERATION).i386.rpm
notifiarr-$(RPMVERSION)-$(ITERATION).i386.rpm: package_build_linux_386_rpm check_fpm
	@echo "Building 32-bit 'rpm' package for notifiarr version '$(RPMVERSION)-$(ITERATION)'."
	fpm -s dir -t rpm $(PACKAGE_ARGS) -a i386 -v $(RPMVERSION) -C $< $(EXTRA_FPM_FLAGS)
	[ "$(SIGNING_KEY)" = "" ] || rpmsign --key-id=$(SIGNING_KEY) --resign notifiarr-$(RPMVERSION)-$(ITERATION).i386.rpm

deb386: notifiarr_$(VERSION)-$(ITERATION)_i386.deb
notifiarr_$(VERSION)-$(ITERATION)_i386.deb: package_build_linux_386_deb check_fpm
	@echo "Building 32-bit 'deb' package for notifiarr version '$(VERSION)-$(ITERATION)'."
	fpm -s dir -t deb $(PACKAGE_ARGS) -a i386 -v $(VERSION) -C $< $(EXTRA_FPM_FLAGS)
	[ "$(SIGNING_KEY)" = "" ] || debsigs --default-key="$(SIGNING_KEY)" --sign=origin notifiarr_$(VERSION)-$(ITERATION)_i386.deb

rpmarm: notifiarr-$(RPMVERSION)-$(ITERATION).aarch64.rpm
notifiarr-$(RPMVERSION)-$(ITERATION).aarch64.rpm: package_build_linux_arm64_rpm check_fpm
	@echo "Building 64-bit ARM8 'rpm' package for notifiarr version '$(RPMVERSION)-$(ITERATION)'."
	fpm -s dir -t rpm $(PACKAGE_ARGS) -a aarch64 -v $(RPMVERSION) -C $< $(EXTRA_FPM_FLAGS)
	[ "$(SIGNING_KEY)" = "" ] || rpmsign --key-id=$(SIGNING_KEY) --resign notifiarr-$(RPMVERSION)-$(ITERATION).aarch64.rpm

debarm: notifiarr_$(VERSION)-$(ITERATION)_arm64.deb
notifiarr_$(VERSION)-$(ITERATION)_arm64.deb: package_build_linux_arm64_deb check_fpm
	@echo "Building 64-bit ARM8 'deb' package for notifiarr version '$(VERSION)-$(ITERATION)'."
	fpm -s dir -t deb $(PACKAGE_ARGS) -a arm64 -v $(VERSION) -C $< $(EXTRA_FPM_FLAGS)
	[ "$(SIGNING_KEY)" = "" ] || debsigs --default-key="$(SIGNING_KEY)" --sign=origin notifiarr_$(VERSION)-$(ITERATION)_arm64.deb

rpmarmhf: notifiarr-$(RPMVERSION)-$(ITERATION).armhf.rpm
notifiarr-$(RPMVERSION)-$(ITERATION).armhf.rpm: package_build_linux_armhf_rpm check_fpm
	@echo "Building 32-bit ARM6/7 HF 'rpm' package for notifiarr version '$(RPMVERSION)-$(ITERATION)'."
	fpm -s dir -t rpm $(PACKAGE_ARGS) -a armhf -v $(RPMVERSION) -C $< $(EXTRA_FPM_FLAGS)
	[ "$(SIGNING_KEY)" = "" ] || rpmsign --key-id=$(SIGNING_KEY) --resign notifiarr-$(RPMVERSION)-$(ITERATION).armhf.rpm

debarmhf: notifiarr_$(VERSION)-$(ITERATION)_armhf.deb
notifiarr_$(VERSION)-$(ITERATION)_armhf.deb: package_build_linux_armhf_deb check_fpm
	@echo "Building 32-bit ARM6/7 HF 'deb' package for notifiarr version '$(VERSION)-$(ITERATION)'."
	fpm -s dir -t deb $(PACKAGE_ARGS) -a armhf -v $(VERSION) -C $< $(EXTRA_FPM_FLAGS)
	[ "$(SIGNING_KEY)" = "" ] || debsigs --default-key="$(SIGNING_KEY)" --sign=origin notifiarr_$(VERSION)-$(ITERATION)_armhf.deb

freebsd_pkg: notifiarr-$(VERSION)_$(ITERATION).amd64.txz
notifiarr-$(VERSION)_$(ITERATION).amd64.txz: package_build_freebsd check_fpm
	@echo "Building 'freebsd pkg' package for notifiarr version '$(VERSION)-$(ITERATION)'."
	fpm -s dir -t freebsd $(PACKAGE_ARGS) -a amd64 -v $(VERSION) -p notifiarr-$(VERSION)_$(ITERATION).amd64.txz -C $< $(EXTRA_FPM_FLAGS)

freebsd386_pkg: notifiarr-$(VERSION)_$(ITERATION).i386.txz
notifiarr-$(VERSION)_$(ITERATION).i386.txz: package_build_freebsd_386 check_fpm
	@echo "Building 32-bit 'freebsd pkg' package for notifiarr version '$(VERSION)-$(ITERATION)'."
	fpm -s dir -t freebsd $(PACKAGE_ARGS) -a 386 -v $(VERSION) -p notifiarr-$(VERSION)_$(ITERATION).i386.txz -C $< $(EXTRA_FPM_FLAGS)

freebsdarm_pkg: notifiarr-$(VERSION)_$(ITERATION).armhf.txz
notifiarr-$(VERSION)_$(ITERATION).armhf.txz: package_build_freebsd_arm check_fpm
	@echo "Building 32-bit ARM6/7 HF 'freebsd pkg' package for notifiarr version '$(VERSION)-$(ITERATION)'."
	fpm -s dir -t freebsd $(PACKAGE_ARGS) -a arm -v $(VERSION) -p notifiarr-$(VERSION)_$(ITERATION).armhf.txz -C $< $(EXTRA_FPM_FLAGS)

# Build an environment that can be packaged for linux.
package_build_linux_rpm: readme man linux
	# Building package environment for linux.
	mkdir -p $@/usr/bin $@/etc/notifiarr $@/usr/share/man/man1 $@/usr/share/doc/notifiarr $@/usr/lib/notifiarr $@/var/log/notifiarr
	# Copying the binary, config file, unit file, and man page into the env.
	cp notifiarr.amd64.linux $@/usr/bin/notifiarr
	cp *.1.gz $@/usr/share/man/man1
	cp examples/notifiarr.conf.example $@/etc/notifiarr/
	cp examples/notifiarr.conf.example $@/etc/notifiarr/notifiarr.conf
	cp LICENSE *.html examples/*?.?* $@/usr/share/doc/notifiarr/
	mkdir -p $@/lib/systemd/system
	cp init/systemd/notifiarr.service $@/lib/systemd/system/
	[ ! -d "init/linux/rpm" ] || cp -r init/linux/rpm/* $@

# Build an environment that can be packaged for linux.
package_build_linux_deb: readme man linux
	# Building package environment for linux.
	mkdir -p $@/usr/bin $@/etc/notifiarr $@/usr/share/man/man1 $@/usr/share/doc/notifiarr $@/usr/lib/notifiarr $@/var/log/notifiarr
	# Copying the binary, config file, unit file, and man page into the env.
	cp notifiarr.amd64.linux $@/usr/bin/notifiarr
	cp *.1.gz $@/usr/share/man/man1
	cp examples/notifiarr.conf.example $@/etc/notifiarr/
	cp examples/notifiarr.conf.example $@/etc/notifiarr/notifiarr.conf
	cp LICENSE *.html examples/*?.?* $@/usr/share/doc/notifiarr/
	mkdir -p $@/lib/systemd/system
	cp init/systemd/notifiarr.service $@/lib/systemd/system/
	[ ! -d "init/linux/deb" ] || cp -r init/linux/deb/* $@

package_build_linux_386_deb: package_build_linux_deb linux386
	mkdir -p $@
	cp -r $</* $@/
	cp notifiarr.386.linux $@/usr/bin/notifiarr

package_build_linux_arm64_deb: package_build_linux_deb arm64
	mkdir -p $@
	cp -r $</* $@/
	cp notifiarr.arm64.linux $@/usr/bin/notifiarr

package_build_linux_armhf_deb: package_build_linux_deb armhf
	mkdir -p $@
	cp -r $</* $@/
	cp notifiarr.arm.linux $@/usr/bin/notifiarr

package_build_linux_386_rpm: package_build_linux_rpm linux386
	mkdir -p $@
	cp -r $</* $@/
	cp notifiarr.386.linux $@/usr/bin/notifiarr

package_build_linux_arm64_rpm: package_build_linux_rpm arm64
	mkdir -p $@
	cp -r $</* $@/
	cp notifiarr.arm64.linux $@/usr/bin/notifiarr

package_build_linux_armhf_rpm: package_build_linux_rpm armhf
	mkdir -p $@
	cp -r $</* $@/
	cp notifiarr.arm.linux $@/usr/bin/notifiarr

# Build an environment that can be packaged for freebsd.
package_build_freebsd: readme man freebsd
	mkdir -p $@/usr/local/bin $@/usr/local/etc/notifiarr $@/usr/local/share/man/man1 $@/usr/local/share/doc/notifiarr $@/usr/local/var/log/notifiarr
	cp notifiarr.amd64.freebsd $@/usr/local/bin/notifiarr
	cp *.1.gz $@/usr/local/share/man/man1
	cp examples/notifiarr.conf.example $@/usr/local/etc/notifiarr/
	cp examples/notifiarr.conf.example $@/usr/local/etc/notifiarr/notifiarr.conf
	cp LICENSE *.html examples/*?.?* $@/usr/local/share/doc/notifiarr/
	mkdir -p $@/usr/local/etc/rc.d
	cp init/bsd/freebsd.rc.d $@/usr/local/etc/rc.d/notifiarr
	chmod +x $@/usr/local/etc/rc.d/notifiarr

package_build_freebsd_386: package_build_freebsd freebsd386
	mkdir -p $@
	cp -r $</* $@/
	cp notifiarr.i386.freebsd $@/usr/local/bin/notifiarr

package_build_freebsd_arm: package_build_freebsd freebsdarm
	mkdir -p $@
	cp -r $</* $@/
	cp notifiarr.armhf.freebsd $@/usr/local/bin/notifiarr

check_fpm:
	@fpm --version > /dev/null || (echo "FPM missing. Install FPM: https://fpm.readthedocs.io/en/latest/installing.html" && false)

# Run code tests and lint.
test: generate lint
	# Testing.
	go test -race -covermode=atomic ./...
lint: generate
	# Checking lint.
	$(shell go env GOPATH)/bin/golangci-lint version
	GOOS=linux $(shell go env GOPATH)/bin/golangci-lint run
	GOOS=freebsd $(shell go env GOPATH)/bin/golangci-lint run
	GOOS=windows $(shell go env GOPATH)/bin/golangci-lint run

generate: pkg/bindata/bindata.go pkg/bindata/docs/api_docs.go
pkg/bindata/docs/api_docs.go: 
	go generate ./pkg/bindata/docs
pkg/bindata/bindata.go:
	find pkg -name .DS\* -delete
	go generate ./pkg/bindata/

##################
##### Docker #####
##################

docker:
	init/docker/makedocker.sh

# Used for Homebrew only. Other distros can create packages.
install: man readme notifiarr
	@echo -  Done Building  -
	@echo -  Local installation with the Makefile is only supported on macOS.
	@echo -  Otherwise, build and install a package: make rpm -or- make deb
	@[ "$(shell uname)" = "Darwin" ] || (echo "Unable to continue, not a Mac." && false)
	@[ "$(PREFIX)" != "" ] || (echo "Unable to continue, PREFIX not set. Use: make install PREFIX=/usr/local ETC=/usr/local/etc" && false)
	@[ "$(ETC)" != "" ] || (echo "Unable to continue, ETC not set. Use: make install PREFIX=/usr/local ETC=/usr/local/etc" && false)
	# Copying the binary, config file, unit file, and man page into the env.
	/usr/bin/install -m 0755 -d $(PREFIX)/bin $(PREFIX)/share/man/man1 $(ETC)/notifiarr $(PREFIX)/share/doc/notifiarr $(PREFIX)/lib/notifiarr
	/usr/bin/install -m 0755 -cp notifiarr $(PREFIX)/bin/notifiarr
	/usr/bin/install -m 0644 -cp notifiarr.1.gz $(PREFIX)/share/man/man1
	/usr/bin/install -m 0644 -cp examples/notifiarr.conf.example $(ETC)/notifiarr/
	[ -f $(ETC)/notifiarr/notifiarr.conf ] || /usr/bin/install -m 0644 -cp  examples/notifiarr.conf.example $(ETC)/notifiarr/notifiarr.conf
	/usr/bin/install -m 0644 -cp LICENSE *.html examples/* $(PREFIX)/share/doc/notifiarr/
