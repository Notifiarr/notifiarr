#!/usr/bin/env bash
# Publish auto-update artifacts to unstable.golift.io with stable names.
#
# Auto-update URLs cannot include a version; it lives in a sibling .txt JSON.
# The payload is a gzipped (or zipped) *binary*, matching the old Makefile
# `gzip -9r` / `zip … $exe` layout — not a tar.gz (gunzip of a tar is a tar).
#
#   Notifiarr.dmg
#   notifiarr.amd64.exe.zip
#   notifiarr.{amd64,386,arm,arm64}.linux.gz
#   notifiarr.{amd64,i386,armhf}.freebsd.gz
#
# Sidecar JSON (pkg/update/unstable.go + userscripts/unstable-syno.sh):
#   {"version":"0.9.8","revision":3273,"size":12345}
# version is the x.y.z prefix of GoReleaser .Version — not the full
# nightly string, or GetUnstable prints version-revision-revision.
set -euo pipefail

if ! command -v jq >/dev/null; then
  echo "jq is required to read artifacts.json" >&2
  exit 1
fi

dir="${1:-dist}"
artifacts="${dir}/artifacts.json"
metadata="${dir}/metadata.json"
combined=""

# Split/merge writes dist/$GOOS/artifacts.json. A single-job release writes dist/artifacts.json.
if [ ! -f "${artifacts}" ]; then
  shopt -s nullglob
  parts=("${dir}"/*/artifacts.json)
  shopt -u nullglob
  if [ ${#parts[@]} -eq 0 ]; then
    echo "missing ${artifacts} and ${dir}/*/artifacts.json" >&2
    find "${dir}" -name artifacts.json -o -name metadata.json >&2 || true
    exit 1
  fi
  combined="$(mktemp "${TMPDIR:-/tmp}/notifiarr-artifacts.XXXXXX")"
  jq -s 'add' "${parts[@]}" > "${combined}"
  artifacts="${combined}"
  echo "merged ${#parts[@]} split artifacts.json files"
fi
if [ ! -f "${metadata}" ]; then
  shopt -s nullglob
  metas=("${dir}"/*/metadata.json)
  shopt -u nullglob
  if [ ${#metas[@]} -gt 0 ]; then
    metadata="${metas[0]}"
  fi
fi

full_version="${VERSION:-}"
if [ -z "${full_version}" ] && [ -f "${metadata}" ]; then
  full_version="$(jq -r '.version // empty' "${metadata}")"
fi
if [ -z "${full_version}" ] || [ "${full_version}" = "unknown" ] || [ "${full_version}" = "unstable" ]; then
  echo "refusing to upload with VERSION=${full_version:-<empty>} (need dist/metadata.json or VERSION=0.9.8-1234)" >&2
  exit 1
fi

# JSON version is x.y.z only; revision is the integer REVISION / nightly suffix.
version="${full_version%%-*}"
revision="${REVISION:-}"
if [ -z "${revision}" ] && [[ "${full_version}" == *-* ]]; then
  revision="${full_version##*-}"
fi
if [[ ! ${revision} =~ ^[0-9]+$ ]]; then
  echo "refusing sidecar revision=${revision:-<empty>} (need REVISION or nightly VERSION suffix)" >&2
  exit 1
fi

stage="${UNSTABLE_STAGE_DIR:-}"
owned_stage=0
if [ -z "${stage}" ]; then
  stage="$(mktemp -d "${TMPDIR:-/tmp}/notifiarr-unstable.XXXXXX")"
  owned_stage=1
fi
if [ "${owned_stage}" -eq 1 ]; then
  trap 'rm -rf "${stage}"; [ -n "${combined}" ] && rm -f "${combined}"' EXIT
elif [ -n "${combined}" ]; then
  trap 'rm -f "${combined}"' EXIT
fi
mkdir -p "${stage}"

resolve_path() {
  local p=$1
  if [ -f "${p}" ]; then
    printf '%s' "${p}"
    return
  fi
  if [ -f "${dir}/${p}" ]; then
    printf '%s' "${dir}/${p}"
    return
  fi
  echo "artifact missing: ${p}" >&2
  return 1
}

# Historical names: linux uses GOARCH; freebsd 386/arm used i386/armhf. No freebsd/arm64.
dest_name() {
  local os=$1 arch=$2
  case "${os}" in
    windows)
      printf 'notifiarr.%s.exe.zip' "${arch}"
      ;;
    linux)
      printf 'notifiarr.%s.linux.gz' "${arch}"
      ;;
    freebsd)
      case "${arch}" in
        386) printf 'notifiarr.i386.freebsd.gz' ;;
        arm) printf 'notifiarr.armhf.freebsd.gz' ;;
        arm64) return 1 ;;
        *) printf 'notifiarr.%s.freebsd.gz' "${arch}" ;;
      esac
      ;;
    *)
      return 1
      ;;
  esac
}

zip_exe() {
  local src=$1 dest=$2 work
  if ! command -v zip >/dev/null; then
    echo "zip is required to stage Windows unstable artifacts" >&2
    exit 1
  fi
  work="$(mktemp -d "${TMPDIR:-/tmp}/notifiarr-zip.XXXXXX")"
  dest="$(cd "$(dirname "${dest}")" && pwd)/$(basename "${dest}")"
  cp -f "${src}" "${work}/notifiarr.exe"
  (cd "${work}" && zip -9q "${dest}" notifiarr.exe)
  rm -rf "${work}"
}

# Prefer GOARM 7 over 6 when both exist (same dest name).
while IFS=$'\t' read -r os arch goarm path; do
  [ -n "${path}" ] || continue
  if [ "${arch}" = arm ] && [ "${goarm}" = 6 ]; then
    continue
  fi
  src="$(resolve_path "${path}")" || exit 1
  dest="$(dest_name "${os}" "${arch}")" || continue
  out="${stage}/${dest}"
  case "${os}" in
    windows)
      zip_exe "${src}" "${out}" </dev/null
      ;;
    *)
      gzip -9nc "${src}" > "${out}" </dev/null
      ;;
  esac
  echo "staged ${dest} from ${src}"
done < <(jq -r '
  .[]
  | select(.type == "Binary")
  | select(.goos == "linux" or .goos == "freebsd" or .goos == "windows")
  | [.goos, .goarch, (.goarm // "-"), .path]
  | @tsv
' "${artifacts}")

# Overlay GoReleaser / legacy_gz Archive names (same files GitHub ships).
# Windows zip then includes LICENSE/docs like the old Makefile zip.
while IFS=$'\t' read -r name path; do
  [ -n "${path}" ] || continue
  base="$(basename "${name}")"
  case "${base}" in
    notifiarr.*.exe.zip|notifiarr.*.linux.gz|notifiarr.*.freebsd.gz)
      src="$(resolve_path "${path}")" || continue
      cp -f "${src}" "${stage}/${base}"
      echo "overlay ${base} from archive ${src}"
      ;;
  esac
done < <(jq -r '.[] | select(.type == "Archive") | [.name, .path] | @tsv' "${artifacts}")

dmg=""
while IFS= read -r path; do
  [ -n "${path}" ] || continue
  src="$(resolve_path "${path}")" || continue
  dmg="${src}"
done < <(jq -r '.[] | select(.type == "DMG") | .path' "${artifacts}")
if [ -z "${dmg}" ]; then
  shopt -s nullglob
  dmgs=("${dir}"/*.dmg)
  shopt -u nullglob
  if [ ${#dmgs[@]} -gt 0 ]; then
    dmg="${dmgs[0]}"
  else
    while IFS= read -r path; do
      [ -n "${path}" ] || continue
      dmg="${path}"
      break
    done < <(find "${dir}" -type f -name '*.dmg' | sort)
  fi
fi
if [ -n "${dmg}" ]; then
  cp -f "${dmg}" "${stage}/Notifiarr.dmg"
  echo "staged Notifiarr.dmg from ${dmg}"
elif [ "${CHANNEL:-}" = nightly ]; then
  echo "warning: no DMG in ${dir} (nightly skips darwin)" >&2
else
  echo "Notifiarr.dmg missing from ${dir}; darwin split/notarize did not produce a DMG" >&2
  exit 1
fi

# Stable URLs keep the previous file if we skip an arch. Require the full set.
required=(
  notifiarr.amd64.exe.zip
  notifiarr.amd64.linux.gz
  notifiarr.386.linux.gz
  notifiarr.arm.linux.gz
  notifiarr.arm64.linux.gz
  notifiarr.amd64.freebsd.gz
  notifiarr.i386.freebsd.gz
  notifiarr.armhf.freebsd.gz
)
if [ "${CHANNEL:-}" != nightly ]; then
  required+=(Notifiarr.dmg)
fi
missing=()
for name in "${required[@]}"; do
  [ -f "${stage}/${name}" ] || missing+=("${name}")
done
if [ ${#missing[@]} -ne 0 ]; then
  echo "unstable upload missing: ${missing[*]}" >&2
  ls -la "${stage}" >&2 || true
  exit 1
fi
shopt -s nullglob
staged=("${stage}"/*)
shopt -u nullglob

upload() {
  local file=$1
  local name size sidecar
  name="$(basename "${file}")"
  size="$(wc -c < "${file}" | tr -d ' ')"
  sidecar="$(mktemp "${TMPDIR:-/tmp}/notifiarr-sidecar.XXXXXX")"
  jq -n --arg version "${version}" --argjson revision "${revision}" --argjson size "${size}" \
    '{version:$version,revision:$revision,size:$size}' > "${sidecar}"
  echo "Uploading ${name} (${version}-${revision}, ${size} bytes)"
  curl -sS --fail-with-body --retry 5 --retry-all-errors --retry-delay 2 \
    -H "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" \
    "https://unstable.golift.io/upload.php?folder=notifiarr" \
    -F "file=@${file};filename=${name}"
  curl -sS --fail-with-body --retry 5 --retry-all-errors --retry-delay 2 \
    -H "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" \
    "https://unstable.golift.io/upload.php?folder=notifiarr" \
    -F "file=@${sidecar};filename=${name}.txt;type=application/json"
  rm -f "${sidecar}"
}

if [ -z "${UNSTABLE_UPLOAD_KEY:-}" ]; then
  if [ -n "${GITHUB_ACTIONS:-}" ]; then
    echo "UNSTABLE_UPLOAD_KEY unset; refusing to skip unstable.golift.io upload in CI" >&2
    exit 1
  fi
  echo "UNSTABLE_UPLOAD_KEY unset; staged without uploading:" >&2
  ls -l "${stage}"
  exit 0
fi

for file in "${staged[@]}"; do
  [ -f "${file}" ] || continue
  upload "${file}"
done
