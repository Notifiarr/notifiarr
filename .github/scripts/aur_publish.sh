#!/usr/bin/env bash
# Push the AUR binary package after a tagged merge.
#
# Keep notifiarr-bin (prebuilt .linux.gz). Do not add GoReleaser aur_sources:
# --split fatals "no linux archives found", merge Publish is a silent no-op.
#
# Usage: aur_publish.sh [dist]
# Env: AUR_DEPLOY_KEY (required to push), VERSION, TAG, CHANNEL, DRY_RUN=1, PKGREL
# AUR is stable vX.Y.Z tags only. pkgrel defaults to 1; re-publish the same
# version with PKGREL bumped or AUR helpers treat it as a downgrade.
set -euo pipefail

dir="${1:-dist}"
pkgname=notifiarr-bin
appname=notifiarr
pkgrel="${PKGREL:-1}"
pkgdesc='Official Client for Notifiarr.com'
url='https://notifiarr.com'
repo="${GITHUB_REPOSITORY:-Notifiarr/notifiarr}"
aur_git='ssh://aur@aur.archlinux.org/notifiarr-bin.git'

if [ "${CHANNEL:-}" = nightly ] || [ "${CHANNEL:-}" = unstable ]; then
  echo "skipping AUR for CHANNEL=${CHANNEL}"
  exit 0
fi

version="${VERSION:-}"
version="${version#v}"
if [ -z "${version}" ] && [[ "${GITHUB_REF:-}" == refs/tags/v* ]]; then
  version="${GITHUB_REF#refs/tags/v}"
fi
if [ -z "${version}" ]; then
  shopt -s nullglob
  metas=("${dir}/metadata.json" "${dir}"/*/metadata.json)
  shopt -u nullglob
  for meta in "${metas[@]}"; do
    [ -f "${meta}" ] || continue
    version="$(jq -r '.version // empty' "${meta}" 2>/dev/null || true)"
    [ -n "${version}" ] && [ "${version}" != unknown ] && [ "${version}" != unstable ] && break
    version=
  done
fi
if [ -z "${version}" ] || [ "${version}" = unknown ] || [ "${version}" = unstable ]; then
  echo "refusing AUR publish with VERSION=${version:-<empty>}" >&2
  exit 1
fi
if [[ ! ${version} =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "refusing AUR pkgver '${version}' (stable x.y.z tags only)" >&2
  exit 1
fi

tag="${TAG:-v${version}}"
rel="https://github.com/${repo}/releases/download/${tag}"
src_url="https://github.com/${repo}/archive/refs/tags/${tag}.tar.gz"
url_amd64="${rel}/notifiarr.amd64.linux.gz"
url_arm="${rel}/notifiarr.arm.linux.gz"
url_arm64="${rel}/notifiarr.arm64.linux.gz"
url_386="${rel}/notifiarr.386.linux.gz"

sha256() {
  if command -v sha256sum >/dev/null; then
    sha256sum "$1" | awk '{print $1}'
  else
    shasum -a 256 "$1" | awk '{print $1}'
  fi
}

find_gz() {
  local name=$1
  local f
  if [ -f "${dir}/linux/${name}" ]; then
    printf '%s\n' "${dir}/linux/${name}"
    return 0
  fi
  if [ -f "${dir}/${name}" ]; then
    printf '%s\n' "${dir}/${name}"
    return 0
  fi
  f="$(find "${dir}" -name "${name}" -type f 2>/dev/null | head -1 || true)"
  if [ -n "${f}" ]; then
    printf '%s\n' "${f}"
    return 0
  fi
  return 1
}

# CI must hash the merge-job gz files (legacy_gz after linux split). Local
# runs may curl the GitHub Release when dist/ is empty (first AUR create).
hash_gz() {
  local name=$1 url=$2
  local f
  if f="$(find_gz "${name}")"; then
    echo "using ${f}" >&2
    sha256 "${f}"
    return 0
  fi
  if [ -n "${GITHUB_ACTIONS:-}" ]; then
    echo "missing ${name} under ${dir}; linux split should have written it via legacy_gz.sh" >&2
    find "${dir}" -name '*.linux.gz' >&2 || ls -l "${dir}" >&2 || true
    exit 1
  fi
  echo "missing ${name}; downloading ${url}" >&2
  mkdir -p "${dir}"
  curl -fsSL -o "${dir}/${name}" "${url}"
  sha256 "${dir}/${name}"
}

# GitHub tag archive is what AUR users download (goreleaser source is a
# different blob). Always hash that URL.
mkdir -p "${dir}"
src_stage="${dir}/aur-src-${version}.tar.gz"
if [ ! -f "${src_stage}" ]; then
  echo "downloading ${src_url}" >&2
  curl -fsSL -o "${src_stage}" "${src_url}"
fi
sum="$(sha256 "${src_stage}")"
sum_amd64="$(hash_gz notifiarr.amd64.linux.gz "${url_amd64}")"
sum_arm="$(hash_gz notifiarr.arm.linux.gz "${url_arm}")"
sum_arm64="$(hash_gz notifiarr.arm64.linux.gz "${url_arm64}")"
sum_386="$(hash_gz notifiarr.386.linux.gz "${url_386}")"
pkgver="${version}"

stage="$(mktemp -d "${TMPDIR:-/tmp}/notifiarr-aur.XXXXXX")"
cleanup() {
  rm -rf "${stage}"
  if [ -n "${keyfile:-}" ]; then
    rm -f "${keyfile}"
  fi
}
trap cleanup EXIT

cp -f init/systemd/notifiarr.install "${stage}/notifiarr.install"

# PKGBUILD functions keep ${pkgname}/${pkgver}/${pkgdir} for makepkg.
cat > "${stage}/PKGBUILD" <<EOF
# Maintainer: David Newhall II <captain at golift dot io>
# Maintainer: Donald Webster <fryfrog at gmail dot com>

pkgname='${pkgname}'
appname='${appname}'
pkgver=${pkgver}
pkgrel=${pkgrel}
pkgdesc='${pkgdesc}'
url='${url}'
arch=('x86_64' 'armhf' 'armv7h' 'aarch64' 'i686' 'pentium4')
license=('MIT')
provides=('notifiarr')
makedepends=('go' 'gzip')
options=('!strip')
backup=('etc/notifiarr/notifiarr.conf')
install=notifiarr.install
source=("\${pkgname}-\${pkgver}.tar.gz::${src_url}")
sha256sums=('${sum}')
source_x86_64=("\${pkgname}-\${pkgver}.x86_64.gz::${url_amd64}")
source_armhf=("\${pkgname}-\${pkgver}.armhf.gz::${url_arm}")
source_armv7h=("\${pkgname}-\${pkgver}.armv7h.gz::${url_arm}")
source_aarch64=("\${pkgname}-\${pkgver}.aarch64.gz::${url_arm64}")
source_i686=("\${pkgname}-\${pkgver}.i686.gz::${url_386}")
source_pentium4=("\${pkgname}-\${pkgver}.pentium4.gz::${url_386}")
sha256sums_x86_64=('${sum_amd64}')
sha256sums_armhf=('${sum_arm}')
sha256sums_armv7h=('${sum_arm}')
sha256sums_aarch64=('${sum_arm64}')
sha256sums_i686=('${sum_386}')
sha256sums_pentium4=('${sum_386}')

build() {
  cd "\${appname}-\${pkgver}"
  go run github.com/davidnewhall/md2roff@v0.0.1 --manual "\${appname}" --version "\${pkgver}" --date "\$(date -u +%Y-%m-%d)" README.md
  go run github.com/davidnewhall/md2roff@v0.0.1 --manual "\${appname}" --version "\${pkgver}" --date "\$(date -u +%Y-%m-%d)" examples/MANUAL.md
  gzip -9nf examples/MANUAL
  mv examples/MANUAL.gz "\${appname}.1.gz"
}

package() {
  install -D -m 755 "\${pkgname}-\${pkgver}.\${CARCH}" "\${pkgdir}/usr/bin/\${appname}"
  cd "\${appname}-\${pkgver}"
  install -d -m 755 "\${pkgdir}/usr/share/licenses/\${appname}" "\${pkgdir}/usr/share/doc/\${appname}" "\${pkgdir}/usr/share/applications" "\${pkgdir}/etc/\${appname}" "\${pkgdir}/var/log/\${appname}"
  install -D -m 644 "examples/\${appname}.conf.example" "\${pkgdir}/etc/\${appname}/\${appname}.conf"
  install -D -m 644 "examples/\${appname}.conf.example" "\${pkgdir}/etc/\${appname}/\${appname}.conf.example"
  install -D -m 644 LICENSE "\${pkgdir}/usr/share/licenses/\${appname}/LICENSE"
  install -D -m 644 examples/MANUAL.html "\${pkgdir}/usr/share/doc/\${appname}/notifiarr_manual.html"
  install -D -m 644 README.html "\${pkgdir}/usr/share/doc/\${appname}/README.html"
  install -D -m 644 examples/compose.yml "\${pkgdir}/usr/share/doc/\${appname}/compose.yml"
  install -D -m 644 examples/prometheus.yml "\${pkgdir}/usr/share/doc/\${appname}/prometheus.yml"
  install -D -m 644 examples/dashboard.json "\${pkgdir}/usr/share/doc/\${appname}/dashboard.json"
  install -D -m 644 "examples/\${appname}.conf.example" "\${pkgdir}/usr/share/doc/\${appname}/\${appname}.conf.example"
  install -D -m 644 "frontend/public/\${appname}.png" "\${pkgdir}/usr/share/doc/\${appname}/\${appname}.png"
  install -D -m 644 "\${appname}.1.gz" "\${pkgdir}/usr/share/man/man1/\${appname}.1.gz"
  install -D -m 644 "init/linux/deb/usr/share/applications/\${appname}.desktop" "\${pkgdir}/usr/share/applications/\${appname}.desktop"
  install -D -m 644 "init/systemd/\${appname}.service" "\${pkgdir}/usr/lib/systemd/system/\${appname}.service"
  echo "u \${appname} - \\"\${appname} daemon\\"" > "\${appname}.sysusers"
  install -D -m 644 "\${appname}.sysusers" "\${pkgdir}/usr/lib/sysusers.d/\${appname}.conf"
  install -D -m 644 "init/systemd/\${appname}.tmpfiles" "\${pkgdir}/usr/lib/tmpfiles.d/\${appname}.conf"
}
EOF

{
  echo "pkgbase = ${pkgname}"
  echo "	pkgdesc = ${pkgdesc}"
  echo "	pkgver = ${pkgver}"
  echo "	pkgrel = ${pkgrel}"
  echo "	url = ${url}"
  echo "	arch = x86_64"
  echo "	arch = armhf"
  echo "	arch = armv7h"
  echo "	arch = aarch64"
  echo "	arch = i686"
  echo "	arch = pentium4"
  echo "	license = MIT"
  echo "	makedepends = go"
  echo "	makedepends = gzip"
  echo "	provides = notifiarr"
  echo "	options = !strip"
  echo "	backup = etc/notifiarr/notifiarr.conf"
  echo "	install = notifiarr.install"
  echo "	source = ${pkgname}-${pkgver}.tar.gz::${src_url}"
  echo "	sha256sums = ${sum}"
  echo "	source_x86_64 = ${pkgname}-${pkgver}.x86_64.gz::${url_amd64}"
  echo "	sha256sums_x86_64 = ${sum_amd64}"
  echo "	source_armhf = ${pkgname}-${pkgver}.armhf.gz::${url_arm}"
  echo "	sha256sums_armhf = ${sum_arm}"
  echo "	source_armv7h = ${pkgname}-${pkgver}.armv7h.gz::${url_arm}"
  echo "	sha256sums_armv7h = ${sum_arm}"
  echo "	source_aarch64 = ${pkgname}-${pkgver}.aarch64.gz::${url_arm64}"
  echo "	sha256sums_aarch64 = ${sum_arm64}"
  echo "	source_i686 = ${pkgname}-${pkgver}.i686.gz::${url_386}"
  echo "	sha256sums_i686 = ${sum_386}"
  echo "	source_pentium4 = ${pkgname}-${pkgver}.pentium4.gz::${url_386}"
  echo "	sha256sums_pentium4 = ${sum_386}"
  echo
  echo "pkgname = ${pkgname}"
} > "${stage}/.SRCINFO"

echo "AUR ${pkgname} ${pkgver}-${pkgrel}"

if [ "${DRY_RUN:-}" = 1 ]; then
  mkdir -p "${dir}/aur"
  cp -f "${stage}/PKGBUILD" "${stage}/.SRCINFO" "${stage}/notifiarr.install" "${dir}/aur/"
  echo "DRY_RUN=1 wrote ${dir}/aur/PKGBUILD"
  exit 0
fi

if [ -z "${AUR_DEPLOY_KEY:-}" ]; then
  if [ -n "${GITHUB_ACTIONS:-}" ]; then
    echo "AUR_DEPLOY_KEY unset; refusing to skip AUR push in CI" >&2
    exit 1
  fi
  echo "AUR_DEPLOY_KEY unset; wrote files without pushing:" >&2
  mkdir -p "${dir}/aur"
  cp -f "${stage}/PKGBUILD" "${stage}/.SRCINFO" "${stage}/notifiarr.install" "${dir}/aur/"
  ls -l "${dir}/aur"
  exit 0
fi

keyfile="$(mktemp "${TMPDIR:-/tmp}/notifiarr-aur-key.XXXXXX")"
printf '%s\n' "${AUR_DEPLOY_KEY}" | tr -d '\r' > "${keyfile}"
chmod 600 "${keyfile}"

export GIT_SSH_COMMAND="ssh -i ${keyfile} -o IdentitiesOnly=yes -o StrictHostKeyChecking=accept-new -F /dev/null"

clone="${stage}/repo"
# Empty AUR packages have no HEAD; --depth 1 fails on first create.
git clone "${aur_git}" "${clone}"
cp -f "${stage}/PKGBUILD" "${stage}/.SRCINFO" "${stage}/notifiarr.install" "${clone}/"
git -C "${clone}" config user.name goreleaserbot
git -C "${clone}" config user.email bot@goreleaser.com
git -C "${clone}" add PKGBUILD .SRCINFO notifiarr.install
if git -C "${clone}" diff --cached --quiet; then
  echo "AUR already at ${pkgver}-${pkgrel}"
  exit 0
fi
git -C "${clone}" commit -m "Update notifiarr-bin to ${tag}"
git -C "${clone}" push origin HEAD:master
echo "pushed AUR notifiarr-bin ${pkgver}-${pkgrel}"
