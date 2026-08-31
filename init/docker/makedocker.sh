#!/usr/bin/env bash

# Local amd64 image from the runtime-only Alpine Dockerfile.
# Releases use GoReleaser Pro dockers_v2 (see .github/workflows/README.md).

set -euo pipefail

source settings.sh

root="$(cd "$(dirname "$0")/../.." && pwd)"
stage="${root}/.docker-build/linux/amd64"
mkdir -p "${stage}"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
  go build -tags osusergo,netgo -trimpath -mod=readonly -modcacherw \
  -o "${stage}/notifiarr" \
  -ldflags "-w -s \
    -X \"golift.io/version.Branch=${BRANCH} (${COMMIT})\" \
    -X \"golift.io/version.BuildDate=${DATE}\" \
    -X \"golift.io/version.BuildUser=$(whoami || echo unknown)\" \
    -X \"golift.io/version.Revision=${ITERATION}\" \
    -X \"golift.io/version.Version=${VERSION}\"" \
  "${root}"

docker buildx build --load --pull --tag notifiarr \
    --progress=plain \
    --platform linux/amd64 \
    --file "${root}/init/docker/Dockerfile.alpine" \
    "${root}/.docker-build"
