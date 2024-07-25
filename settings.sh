MAINT="David Newhall II <captain at golift dot io>"
DESC="Official Client for Notifiarr.com"
LICENSE="MIT"
# Used for source links in package metadata and docker labels.
SOURCE_URL="https://github.com/Notifiarr/notifiarr"
VENDOR="Go Lift <code@golift.io>"
export MAINT DESC LICENSE SOURCE_URL VENDOR

DATE="$(date -u +%Y-%m-%dT%H:%M:00Z)"
VERSION=$(git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1 2>/dev/null) 2>/dev/null | tr -d v)
[ "$VERSION" != "" ] || VERSION=development
# This produces a 0 in some environments (like Homebrew), but it's only used for packages.
ITERATION=$(git rev-list --count --all 2>/dev/null || echo 0)
COMMIT="$(git rev-parse --short HEAD 2>/dev/null || echo 0)"
GIT_BRANCH="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo unknown)"
BRANCH="${GIT_BRANCH:-${GITHUB_REF_NAME}}"
export DATE VERSION ITERATION COMMIT BRANCH

### Optional ###

# Import this signing key only if it's in the keyring.
if gpg --list-keys 2>/dev/null | grep -q B93DD66EF98E54E2EAE025BA0166AD34ABC5A57C; then
    export SIGNING_KEY=B93DD66EF98E54E2EAE025BA0166AD34ABC5A57C
fi

# Make sure Docker builds work locally.
# These do not affect automated builds, just allow the docker build scripts to run from a local clone.
[ -n "$SOURCE_BRANCH" ] || export SOURCE_BRANCH=$BRANCH
[ -n "$DOCKER_TAG" ] || export DOCKER_TAG=$(echo $SOURCE_BRANCH | sed 's/^v*\([0-9].*\)/\1/')
[ -n "$DOCKER_REPO" ] || export DOCKER_REPO="golift/notifiarr"
[ -n "$IMAGE_NAME" ] || export IMAGE_NAME="${DOCKER_REPO}:${DOCKER_TAG}"
[ -n "$DOCKERFILE_PATH" ] || export DOCKERFILE_PATH="init/docker/Dockerfile.scratch"
