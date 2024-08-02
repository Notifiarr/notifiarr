#!/bin/bash -x

# Deploys a new aur PKGBUILD file to the Arch Linux AUR repo.
# Run by GitHub Actions when a new release is created on GitHub.

source settings.sh

sha512sum=sha512sum
$sha512sum -v 2>/dev/null || sha512sum="shasum -a 512" # macos

SOURCE_PATH="https://github.com/Notifiarr/notifiarr/archive/v${VERSION}.tar.gz"
echo "==> Using URL: $SOURCE_PATH"
SHA=$(curl -sL "$SOURCE_PATH" | $sha512sum | awk '{print $1}')

if [ -f sha512sums.txt ]; then
  echo "==> Using SHA512SUMS file"
  SHA_X64=$(grep notifiarr.amd64.linux.gz sha512sums.txt | awk '{print $1}')
  SHA_ARMHF=$(grep notifiarr.arm.linux.gz sha512sums.txt | awk '{print $1}')
  SHA_ARCH64=$(grep notifiarr.arm64.linux.gz sha512sums.txt | awk '{print $1}')
  SHA_386=$(grep notifiarr.386.linux.gz sha512sums.txt | awk '{print $1}')
else
  echo "==> Not Using SHA512SUMS file"
  source_x64="https://github.com/Notifiarr/notifiarr/releases/download/v${VERSION}/notifiarr.amd64.linux.gz"
  source_armhf="https://github.com/Notifiarr/notifiarr/releases/download/v${VERSION}/notifiarr.arm.linux.gz"
  source_arm64="https://github.com/Notifiarr/notifiarr/releases/download/v${VERSION}/notifiarr.arm64.linux.gz"
  source_386="https://github.com/Notifiarr/notifiarr/releases/download/v${VERSION}/notifiarr.386.linux.gz"
  SHA_X64=$(curl -sL "$source_x64" | $sha512sum | awk '{print $1}')
  SHA_ARMHF=$(curl -sL "$source_armhf" | $sha512sum | awk '{print $1}')
  SHA_ARCH64=$(curl -sL "$source_arm64" | $sha512sum | awk '{print $1}')
  SHA_386=$(curl -sL "$source_386" | $sha512sum | awk '{print $1}')
fi

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

set -e

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
