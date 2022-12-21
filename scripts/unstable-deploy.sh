#!/usr/bin/env bash

source settings.sh

URL="https://unstable.notifiarr.app/upload.php"

if [ -z "${UNSTABLE_UPLOAD_KEY}" ]; then
  echo "No upload key."
  exit 0
fi

for file in release/*.{zip,dmg,gz,txz}; do
  if [ -f "$file" ]; then
    curl -H "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" "$URL" -F "file=@$file"
  fi
done