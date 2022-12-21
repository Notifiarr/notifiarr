#!/usr/bin/env bash

if [ -z "${UNSTABLE_UPLOAD_KEY}" ]; then
  echo "No upload key."
  exit 0
fi

[ -f releases/Notifiarr.dmg ] &&
    curl -H "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" https://unstable.notifiarr.app/upload.php -F "file=@release/Notifiarr.dmg"
[ -f releases/Notifiarr-unsigned.dmg ] &&
    curl -H "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" https://unstable.notifiarr.app/upload.php -F "file=@release/Notifiarr-unsigned.dmg"
[ -f releases/notifiarr.amd64.gz ] &&
    curl -H "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" https://unstable.notifiarr.app/upload.php -F "file=@release/notifiarr.amd64.gz"
[ -f releases/notifiarr.amd64.exe ] &&
    curl -H "X-API-KEY: ${UNSTABLE_UPLOAD_KEY}" https://unstable.notifiarr.app/upload.php -F "file=@release/notifiarr.amd64.exe"