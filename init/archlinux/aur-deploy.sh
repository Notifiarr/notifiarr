#!/bin/bash -x

# Deploys a new aur PKGBUILD file to the Arch Linux AUR repo.
# Run by GitHub Actions on tagged releases after GoReleaser merge.

source settings.sh

if [[ -n ${REVISION:-} ]]; then
  ITERATION="${REVISION}"
fi

sha512sum=sha512sum
$sha512sum -v 2>/dev/null || sha512sum="shasum -a 512" # macos

set -euo pipefail

hash_file() {
  $sha512sum "$1" | awk '{print $1}'
}

find_gz() {
  find dist -name "$1" -print -quit 2>/dev/null
}

hash_or_fetch() {
  local name=$1 url=$2
  local f
  f=$(find_gz "${name}")
  if [[ -n $f && -f $f ]]; then
    echo "==> Hashing ${f}" >&2
    hash_file "$f"
  else
    echo "==> Fetching ${url}" >&2
    curl -sS --fail-with-body "$url" | $sha512sum | awk '{print $1}'
  fi
}

SOURCE_PATH="https://github.com/Notifiarr/notifiarr/archive/v${VERSION}.tar.gz"
echo "==> Using URL: $SOURCE_PATH"
SHA=$(curl -sS --fail-with-body "$SOURCE_PATH" | $sha512sum | awk '{print $1}')

SHA_X64=$(hash_or_fetch notifiarr.amd64.linux.gz "https://github.com/Notifiarr/notifiarr/releases/download/v${VERSION}/notifiarr.amd64.linux.gz")
SHA_ARMHF=$(hash_or_fetch notifiarr.arm.linux.gz "https://github.com/Notifiarr/notifiarr/releases/download/v${VERSION}/notifiarr.arm.linux.gz")
SHA_ARCH64=$(hash_or_fetch notifiarr.arm64.linux.gz "https://github.com/Notifiarr/notifiarr/releases/download/v${VERSION}/notifiarr.arm64.linux.gz")
SHA_386=$(hash_or_fetch notifiarr.386.linux.gz "https://github.com/Notifiarr/notifiarr/releases/download/v${VERSION}/notifiarr.386.linux.gz")

push_it() {
  git config user.email "code@golift.io"
  git config user.name "notifiarr-github-releaser"
  pushd release_repo
  git add .
  git commit -m "Update notifiarr on Release: v${VERSION}-${ITERATION}"
  git push
  popd
  rm -rf release_repo
}

if [[ -n $DEPLOY_KEY ]]; then
  mkdir "${HOME}/.ssh/"
  KEY_FILE=$(mktemp -u "${HOME}/.ssh/XXXXX")
  echo "${DEPLOY_KEY}" > "${KEY_FILE}"
  chmod 600 "${KEY_FILE}"
  # Configure ssh to use this secret.
  export GIT_SSH_COMMAND="ssh -i ${KEY_FILE} -o 'StrictHostKeyChecking no'"
fi

rm -rf release_repo
git clone aur@aur.archlinux.org:notifiarr-bin.git release_repo

sed -e "s/{{VERSION}}/${VERSION}/g" \
    -e "s/{{Iter}}/${ITERATION}/g" \
    -e "s/{{SHA}}/${SHA}/g" \
    -e "s/{{Desc}}/${DESC}/g" \
    -e "s%{{SHA_X64}}%${SHA_X64}%g" \
    -e "s%{{SHA_ARMHF}}%${SHA_ARMHF}%g" \
    -e "s%{{SHA_ARCH64}}%${SHA_ARCH64}%g" \
    -e "s%{{SHA_386}}%${SHA_386}%g" \
    init/archlinux/PKGBUILD.template | tee release_repo/PKGBUILD

sed -e "s/{{VERSION}}/${VERSION}/g" \
    -e "s/{{Iter}}/${ITERATION}/g" \
    -e "s/{{SHA}}/${SHA}/g" \
    -e "s/{{Desc}}/${DESC}/g" \
    -e "s%{{SHA_X64}}%${SHA_X64}%g" \
    -e "s%{{SHA_ARMHF}}%${SHA_ARMHF}%g" \
    -e "s%{{SHA_ARCH64}}%${SHA_ARCH64}%g" \
    -e "s%{{SHA_386}}%${SHA_386}%g" \
    init/archlinux/SRCINFO.template | tee release_repo/.SRCINFO

[ "$1" != "" ] || push_it
