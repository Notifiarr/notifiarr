#!/usr/bin/env bash

set -e -o pipefail

# Authenticode-sign a Windows PE via golift/codesign (YubiKey-backed signerd).
# GitHub Actions uses golift/codesign@v1 after `make release WINDOWS_ZIP=0`.
# This script is for local `make windows` when CODESIGN_URL is set (SSH tunnel
# or a configured CLI). PKCS#12 secrets are retired.
#
# On macOS, never call /usr/bin/codesign (Apple's tool). Prefer CODESIGN_BIN
# or "$(go env GOPATH)/bin/codesign".

function pick_codesign() {
  if [ -n "${CODESIGN_BIN:-}" ]; then
    echo "${CODESIGN_BIN}"
    return
  fi
  gopath="$(go env GOPATH 2>/dev/null || true)"
  if [ -n "${gopath}" ] && [ -x "${gopath}/bin/codesign" ]; then
    echo "${gopath}/bin/codesign"
    return
  fi
  case "$(uname -s)" in
    Darwin)
      return 1
      ;;
    *)
      command -v codesign
      ;;
  esac
}

function sign() {
  if [ -z "${CODESIGN_URL:-}" ]; then
    echo "Skipped signing ${FILE} (CODESIGN_URL unset) .." >&2
    exit 0
  fi

  bin="$(pick_codesign)" || {
    echo "CODESIGN_URL is set but golift codesign CLI not found (set CODESIGN_BIN)" >&2
    exit 1
  }

  CODESIGN_NAME="${CODESIGN_NAME:-Notifiarr}" \
  CODESIGN_WEBSITE="${CODESIGN_WEBSITE:-https://notifiarr.com}" \
  "${bin}" -- "${FILE}"
  echo "Signed ${FILE} .." >&2
}

[ -z "$1" ] || FILE="$1" sign
