#!/usr/bin/env bash
# After goreleaser --split, turn freebsd binaries into pkgng .txz (nFPM cannot).
# Same tool as v0.9.7: fpm -s dir -t freebsd. This only stages files, calls fpm,
# and lists the archives in artifacts.json for merge.
# FreeBSD arches match the auto-updater / historical GitHub set: amd64, i386, armhf.
set -euo pipefail

DIST=${1:?usage: freebsd_txz.sh dist/freebsd}
REPO=$(cd "$(dirname "$0")/../.." && pwd)
DIST=$(cd "${DIST}" && pwd)
command -v fpm >/dev/null && command -v jq >/dev/null || { echo "need fpm and jq" >&2; exit 1; }

# fpm freebsd.rb runs `tar --transform` (GNU). Put gtar first on macOS.
if ! tar --version 2>/dev/null | grep -q 'GNU tar'; then
  command -v gtar >/dev/null || { echo "fpm -t freebsd needs GNU tar" >&2; exit 1; }
  _gnutar_bin=$(mktemp -d)
  ln -s "$(command -v gtar)" "${_gnutar_bin}/tar"
  export PATH="${_gnutar_bin}:${PATH}"
fi

artifacts=${DIST}/artifacts.json
VERSION=$(jq -r '.version // empty' "${DIST}/metadata.json")
if [[ -z ${VERSION} || ${VERSION} == unknown || ${VERSION} == unstable ]]; then
  echo "refusing version '${VERSION}'" >&2
  exit 1
fi
# Tagged pkgng version is 0.9.8_REVISION (same as old fpm). --nightly already
# puts REVISION in metadata.json (.Version).
if [[ ${VERSION} =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  [[ -n ${REVISION:-} ]] || { echo "REVISION required for tagged freebsd pkg" >&2; exit 1; }
  VERSION="${VERSION}_${REVISION}"
fi

# goarch -> asset suffix / fpm -a / pkg ABI CPU (uname -p).
pkg_for() {
  case $1 in
    amd64) echo amd64 amd64 amd64 ;;
    386) echo i386 i386 i386 ;;
    arm) echo armhf amd64 armv7 ;;
    *) echo "unsupported freebsd goarch $1" >&2; return 1 ;;
  esac
}

CONF=/usr/local/etc/notifiarr/notifiarr.conf
# fpm --config-files is ignored by freebsd.rb. Unpack only when ABI or config still needs a patch.
fix_pkg() {
  local pkg=$1 abi=$2 tmp list got cfg
  got=$(tar -xJOf "${pkg}" +COMPACT_MANIFEST | jq -r .arch)
  cfg=$(tar -xJOf "${pkg}" +MANIFEST | jq -r '.config[0] // empty')
  [[ ${got} == "${abi}" && ${cfg} == "${CONF}" ]] && return
  tmp=$(mktemp -d)
  tar --transform 's|^/||' -xJf "${pkg}" -C "${tmp}"
  jq --arg a "${abi}" '.arch=$a' "${tmp}/+COMPACT_MANIFEST" > "${tmp}/c.json"
  jq --arg a "${abi}" --arg c "${CONF}" '.arch=$a | .config=[$c]' "${tmp}/+MANIFEST" > "${tmp}/m.json"
  mv "${tmp}/c.json" "${tmp}/+COMPACT_MANIFEST"
  mv "${tmp}/m.json" "${tmp}/+MANIFEST"
  list=$(mktemp)
  # fpm lists files and empty dirs as leaves. A files-only walk drops
  # /usr/local/var/log/notifiarr; the rc script cannot create it as notifiarr.
  { printf '%s\n' +COMPACT_MANIFEST +MANIFEST; (cd "${tmp}" && find usr \( -type f -o -type d -empty \) | sort); } > "${list}"
  tar --owner=0 --group=0 --numeric-owner --no-recursion -Jcf "${pkg}" -C "${tmp}" \
    --files-from "${list}" --transform 's|^\([^+]\)|/\1|'
  rm -rf "${tmp}" "${list}"
}

stage() {
  local bin=$1 root=$2 doc=${2}/usr/local/share/doc/notifiarr
  mkdir -p "${root}/usr/local/bin" "${root}/usr/local/etc/notifiarr" \
    "${root}/usr/local/etc/rc.d" "${root}/usr/local/share/man/man1" \
    "${root}/usr/local/var/log/notifiarr" "${doc}"
  install -m 755 "${bin}" "${root}/usr/local/bin/notifiarr"
  install -m 755 "${REPO}/init/bsd/freebsd.rc.d" "${root}/usr/local/etc/rc.d/notifiarr"
  install -m 644 "${REPO}/examples/notifiarr.conf.example" "${root}/usr/local/etc/notifiarr/notifiarr.conf"
  install -m 644 "${REPO}/examples/notifiarr.conf.example" "${root}/usr/local/etc/notifiarr/notifiarr.conf.example"
  [[ -f ${REPO}/notifiarr.1.gz ]] || { echo "missing notifiarr.1.gz" >&2; exit 1; }
  install -m 644 "${REPO}/notifiarr.1.gz" "${root}/usr/local/share/man/man1/notifiarr.1.gz"
  install -m 644 "${REPO}/LICENSE" "${doc}/LICENSE"
  install -m 644 "${REPO}/examples/MANUAL.md" "${doc}/MANUAL.md"
  [[ -f ${REPO}/examples/MANUAL.html ]] && install -m 644 "${REPO}/examples/MANUAL.html" "${doc}/notifiarr_manual.html"
  [[ -f ${REPO}/examples/compose.yml ]] && install -m 644 "${REPO}/examples/compose.yml" "${doc}/compose.yml"
  install -m 644 "${REPO}/examples/notifiarr.conf.example" "${doc}/notifiarr.conf.example"
  [[ -f ${REPO}/README.html ]] && install -m 644 "${REPO}/README.html" "${doc}/README.html"
  [[ -f ${REPO}/frontend/public/notifiarr.png ]] && install -m 644 "${REPO}/frontend/public/notifiarr.png" "${doc}/notifiarr.png"
}

extra=$(mktemp)
echo '[]' > "${extra}"
seen=

while IFS='|' read -r goarch goarm path; do
  [[ ${goarch} == arm && ${goarm} == 6 ]] && continue
  [[ ${seen} == *"|${goarch}|"* ]] && continue
  seen+="|${goarch}|"

  bin=${path}
  [[ -f ${bin} ]] || bin=${DIST}/${path}
  [[ -f ${bin} ]] || { echo "freebsd binary missing: ${path}" >&2; exit 1; }

  read -r suffix fpm_a cpu <<<"$(pkg_for "${goarch}")"
  dest=${DIST}/notifiarr-${VERSION}.${suffix}.txz
  tmp=$(mktemp -d)
  stage "${bin}" "${tmp}"
  rm -f "${dest}"
  fpm -s dir -t freebsd --name notifiarr -v "${VERSION}" -a "${fpm_a}" \
    --license MIT --url https://notifiarr.com \
    --maintainer "David Newhall II <captain at golift dot io>" \
    --description "Official Client for Notifiarr.com" \
    --freebsd-origin https://github.com/Notifiarr/notifiarr --freebsd-osversion '*' \
    --before-install "${REPO}/init/systemd/before-install.sh" \
    --after-install "${REPO}/init/systemd/after-install.sh" \
    --before-remove "${REPO}/init/systemd/before-remove.sh" \
    --config-files "${CONF}" \
    -C "${tmp}" -p "${dest}" .
  rm -rf "${tmp}"
  fix_pkg "${dest}" "FreeBSD:*:${cpu}"
  echo "wrote ${dest##*/}" >&2
  # Merge unmarshals Type from internal_type (int), not type (string).
  # UploadableArchive = 1. Path must be repo-relative so checksum/GitHub can open it.
  relpath=${dest#"${REPO}/"}
  jq --arg name "${dest##*/}" --arg path "${relpath}" --arg goarch "${goarch}" \
    --arg goarm "${goarm}" --argjson itype 1 \
    '. + [{name:$name, path:$path, goos:"freebsd", goarch:$goarch,
      goarm: (if $goarm == "" then null else $goarm end),
      type:"Archive", internal_type:$itype,
      extra:{ID:"freebsd-pkg", Format:"txz", Ext:".txz"}}]' \
    "${extra}" > "${extra}.n" && mv "${extra}.n" "${extra}"
done < <(jq -r '.[] | select(.type=="Binary" and .goos=="freebsd") | [.goarch, (.goarm // ""), .path] | join("|")' "${artifacts}")

for a in amd64 386 arm; do
  [[ ${seen} == *"|${a}|"* ]] || { echo "freebsd txz missing goarch ${a}" >&2; exit 1; }
done

jq --slurpfile extra "${extra}" '. + $extra[0]' "${artifacts}" > "${artifacts}.n"
mv "${artifacts}.n" "${artifacts}"
rm -f "${extra}"
echo "appended txz archives to ${artifacts}" >&2
