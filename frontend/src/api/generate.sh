#!/bin/sh
# This vendors all the notifiarr dependencies so we can parse their docs.

set -e
# This variable is used by the generate.go file.
export VENDOR_DIR=../../../vendor
[ ! -d "${VENDOR_DIR}" ] || EXISTS="true"

# Download the dependencies.
echo "$(date "+%Y/%m/%d %H:%M:%S") ==> downloading dependencies"
go mod vendor

# Parse the docs and generate the typescript interfaces.
echo "$(date "+%Y/%m/%d %H:%M:%S") ==> starting generator"
go run .

# Remove the vendor folder if it didn't exist before we started.
[ "$EXISTS" = "true" ] || \
    echo "$(date "+%Y/%m/%d %H:%M:%S") ==> removing vendor folder" && \
    rm -rf "${VENDOR_DIR}"

echo "$(date "+%Y/%m/%d %H:%M:%S") ==> done"
