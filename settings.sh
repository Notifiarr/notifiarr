# Application Builder Configuration File. Customized for: hello-world
# Each line must have an export clause.
# This file is parsed and sourced by the Makefile, Docker and Homebrew builds.
# Powered by Application Builder: https://github.com/golift/application-builder

# Bring in dynamic repo/pull/source info.
source $(dirname "${BASH_SOURCE[0]}")/init/buildinfo.sh

# Must match the repo name to make things easy. Otherwise, fix some other paths.
BINARY="notifiarr"
# github username / repo name
REPO="Notifiarr/notifiarr"
# Github repo containing homebrew formula repo.
HBREPO="golift/homebrew-mugs"
AUREPO="golift/aur"
MAINT="David Newhall II <captain at golift dot io>"
DESC="Unified Client for Notifiarr.com"
GOLANGCI_LINT_ARGS="--enable-all -D exhaustivestruct,nlreturn,interfacer,maligned,scopelint,golint,varnamelen,exhaustruct,nonamedreturns"
# Example must exist at examples/$CONFIG_FILE.example
CONFIG_FILE="notifiarr.conf"
LICENSE="MIT"
# FORMULA is either 'service' or 'tool'. Services run as a daemon, tools do not.
# This affects the homebrew formula (launchd) and linux packages (systemd).
FORMULA="service"

# Used for source links in package metadata and docker labels.
SOURCE_URL="https://github.com/${REPO}"

# This parameter is passed in as -X to go build. Used to override the Version variable in a package.
# This makes a path like golift.io/version.Version=1.3.3
# Name the Version-containing library the same as the github repo, without dashes.
VERSION_PATH="golift.io/version"

# Used by homebrew and arch linux downloads.
SOURCE_PATH=https://codeload.github.com/${REPO}/tar.gz/refs/tags/v${VERSION}

export BINARY HBREPO MAINT VENDOR DESC GOLANGCI_LINT_ARGS CONFIG_FILE
export LICENSE FORMULA SOURCE_URL VERSION_PATH SOURCE_PATH

### Optional ###

# Import this signing key only if it's in the keyring.
gpg --list-keys 2>/dev/null | grep -q B93DD66EF98E54E2EAE025BA0166AD34ABC5A57C
[ "$?" != "0" ] || export SIGNING_KEY=B93DD66EF98E54E2EAE025BA0166AD34ABC5A57C

export WINDOWS_LDFLAGS=""
export MACAPP="Notifiarr"
export EXTRA_FPM_FLAGS="--conflicts=discordnotifier-client>0.0.1 --provides=notifiarr --provides=discordnotifier-client"

# Make sure Docker builds work locally.
# These do not affect automated builds, just allow the docker build scripts to run from a local clone.
[ -n "$SOURCE_BRANCH" ] || export SOURCE_BRANCH=$BRANCH
[ -n "$DOCKER_TAG" ] || export DOCKER_TAG=$(echo $SOURCE_BRANCH | sed 's/^v*\([0-9].*\)/\1/')
[ -n "$DOCKER_REPO" ] || export DOCKER_REPO="golift/${BINARY}"
[ -n "$IMAGE_NAME" ] || export IMAGE_NAME="${DOCKER_REPO}:${DOCKER_TAG}"
[ -n "$DOCKERFILE_PATH" ] || export DOCKERFILE_PATH="init/docker/Dockerfile"
