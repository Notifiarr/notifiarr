#!/usr/bin/env bash

source settings.sh

URL="https://unstable.notifiarr.app/upload.php"

if [ -z "${UNSTABLE_UPLOAD_KEY}" ]; then
  echo "No upload key for unstable."
  exit 0
fi

echo "Uploading unstable files."

for file in release/*.{zip,dmg,gz}; do
  if [ -f "$file" ]; then
    echo "Uploading: $file"
    curl -sSH "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" "$URL" -F "file=@$file"
    versionfile="$VERSION-$ITERATION;filename=$(basname $file).txt;type=text/plain"
    curl -sSH "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" "$URL" -F "file=$versionfile"
  fi
done
