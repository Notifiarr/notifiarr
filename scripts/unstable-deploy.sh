#!/usr/bin/env bash

source settings.sh

URL="https://unstable.notifiarr.app/upload.php"

if [ -z "${UNSTABLE_UPLOAD_KEY}" ]; then
  echo "No upload key for unstable."
  exit 0
fi

curl --version
echo
echo "Uploading unstable files."

for file in release/*.{zip,dmg,gz,txz}; do
  if [ -f "$file" ]; then
    echo "Uploading: $file"
    curl -H "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" "$URL" -XPOST -F "file=@$file;filename=$(basename $file)"
    echo
  fi
done