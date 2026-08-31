AUR package is **notifiarr-bin** (prebuilt Linux `.linux.gz` binaries plus the GitHub tag source tree for man pages and docs).

GitHub Actions publishes it from [`.github/scripts/aur_publish.sh`](../../.github/scripts/aur_publish.sh) on tagged `CHANNEL=release` after packagecloud. Do not add GoReleaser `aur_sources`: `--split` has no source tarball yet, merge Publish is a silent no-op, and a source PKGBUILD cannot `npm run build` without `FONTAWESOME_PACKAGE_TOKEN`.
