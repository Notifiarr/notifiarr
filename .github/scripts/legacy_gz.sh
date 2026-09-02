#!/usr/bin/env bash
# After goreleaser --split, gzip linux/freebsd binaries under historical names
# so GitHub Releases still has notifiarr.amd64.linux.gz (AUR notifiarr-bin).
set -euo pipefail

DIST=${1:?usage: legacy_gz.sh dist/linux|dist/freebsd}
OS=${2:?usage: legacy_gz.sh dist/<dir> linux|freebsd}
REPO=$(cd "$(dirname "$0")/../.." && pwd)
DIST=$(cd "${DIST}" && pwd)
command -v gzip >/dev/null && command -v jq >/dev/null || { echo "need gzip and jq" >&2; exit 1; }

artifacts=${DIST}/artifacts.json
[[ -f ${artifacts} ]] || { echo "missing ${artifacts}" >&2; exit 1; }

dest_name() {
  local arch=$1
  case "${OS}" in
    linux)
      printf 'notifiarr.%s.linux.gz' "${arch}"
      ;;
    freebsd)
      case "${arch}" in
        386) printf 'notifiarr.i386.freebsd.gz' ;;
        arm) printf 'notifiarr.armhf.freebsd.gz' ;;
        *) printf 'notifiarr.%s.freebsd.gz' "${arch}" ;;
      esac
      ;;
    *)
      echo "unsupported os ${OS}" >&2
      return 1
      ;;
  esac
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
  [[ -f ${bin} ]] || { echo "${OS} binary missing: ${path}" >&2; exit 1; }

  name=$(dest_name "${goarch}")
  dest=${DIST}/${name}
  gzip -9nc "${bin}" > "${dest}"
  echo "wrote ${name}" >&2
  relpath=${dest#"${REPO}/"}
  jq --arg name "${name}" --arg path "${relpath}" --arg goarch "${goarch}" \
    --arg goarm "${goarm}" --argjson itype 1 \
    '. + [{name:$name, path:$path, goos:"'"${OS}"'", goarch:$goarch,
      goarm: (if $goarm == "" then null else $goarm end),
      type:"Archive", internal_type:$itype,
      extra:{ID:"legacy-gz", Format:"gz", Ext:".gz"}}]' \
    "${extra}" > "${extra}.n" && mv "${extra}.n" "${extra}"
done < <(jq -r --arg os "${OS}" \
  '.[] | select(.type=="Binary" and .goos==$os) | [.goarch, (.goarm // ""), .path] | join("|")' \
  "${artifacts}")

required=()
case "${OS}" in
  linux) required=(amd64 386 arm arm64) ;;
  freebsd) required=(amd64 386 arm) ;;
esac
for a in "${required[@]}"; do
  [[ ${seen} == *"|${a}|"* ]] || { echo "${OS} gz missing goarch ${a}" >&2; exit 1; }
done

jq --slurpfile extra "${extra}" '. + $extra[0]' "${artifacts}" > "${artifacts}.n"
mv "${artifacts}.n" "${artifacts}"
rm -f "${extra}"
echo "appended ${OS} gz archives to ${artifacts}" >&2
