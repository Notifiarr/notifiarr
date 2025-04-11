#!/usr/bin/env bash

set -e -o pipefail

# https://blog.synapp.nz/2017/06/16/code-signing-a-windows-application-on-linux-on-windows/
# https://stackoverflow.com/a/29073957/1142
function sign() {
  if [ -z "${EXE_SIGNING_KEY}" ] || [ -z "${EXE_SIGNING_KEY_PASSWORD}" ]; then
    echo "Skipped signing ${FILE} .." >&2
    exit 0
  fi

  rm -f "${FILE}.signed"
  echo "${EXE_SIGNING_KEY}" | base64 -d | \
  osslsigncode sign -pkcs12 /dev/stdin \
    -pass "${EXE_SIGNING_KEY_PASSWORD}" \
    -n "Notifiarr" \
    -i "https://notifiarr.com" \
    -t "http://timestamp.comodoca.com/authenticode" \
    -in "${FILE}" -out "${FILE}.signed" \
    && cp "${FILE}.signed" "${FILE}" >> /tmp/pwd >&2 \
    && echo "Signed ${FILE} .." >&2
}

[ -z "$1" ] || FILE="$1" sign
