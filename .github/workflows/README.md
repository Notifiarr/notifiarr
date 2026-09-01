# GitHub Actions

Two workflows. `test-and-lint` (`codetests.yml`) runs tests and golangci-lint on pull requests. `build-and-release` (`release.yml`) is the only publisher.

## Channels

`release.yml` maps the GitHub event to a `CHANNEL` env that `.goreleaser.yaml` reads (`dockers_v2.disable`, packagecloud repo, unstable upload). `--nightly` is a GoReleaser flag: it bumps the version and turns off GitHub Releases.

| Trigger | CHANNEL | GoReleaser extra | What it publishes |
|---|---|---|---|
| Push tag `v*` | `release` | (none) | GitHub Release, Docker `:latest` + version tags, AUR `notifiarr-bin`, packagecloud `golift/pkgs` |
| Push branch `unstable` | `unstable` | `--nightly` | Docker `:unstable` (+ `-cuda` / `-ubuntu`), packagecloud `golift/unstable`, [unstable.golift.io](https://unstable.golift.io/?dir=notifiarr) |
| Cron `33 12 * * *` UTC, or `workflow_dispatch` on `main` | `nightly` | `--nightly` | Docker `:nightly` only |

`workflow_dispatch` on any other ref is refused. Tagged and unstable publishes are **push** only.

`unstable` is a **manual publish branch**. Recut it by pushing the commit you want:

```bash
git push origin main:unstable
```

Do not fast-forward `unstable` from `main` in CI. Calendar nightly builds `main` and does not touch the git `unstable` branch.

Nightly skips the Darwin job. Apple notarization is slow and unused for a Docker-only cut.

## Split, then merge

GoReleaser Pro `--split` / `--continue --merge` builds each GOOS in its own job, then one merge job publishes. That exists because:

- Darwin needs **CGO** (`energye/systray` Cocoa) and **native** `codesign` / `notarytool`. Quill on Linux can sign a naked binary; a `.app` inside a DMG is rejected by Gatekeeper unless the bundle is signed on macOS.
- Windows Authenticode talks to house **signerd** (`golift.io/codesign`) and needs `id-token: write` for GitHub OIDC. That is ubuntu, not macOS.
- Linux nFPM (deb/rpm) needs `rpm` + GPG. FreeBSD pkgng `.txz` is built after `--split` by `fpm -t freebsd` (nFPM has no freebsd target). Wrapper: `.github/scripts/freebsd_txz.sh`.
- Every split compiles the frontend (`go generate ./frontend`), so each job has Node 24. Icons are public `phosphor-svelte`; there is no private npm token.

So:

1. **channel** — compute `CHANNEL`, extra args, and `REVISION` (`git rev-list --count --all`). This is the only place the count is taken; later jobs pass `needs.channel.outputs.revision`.
2. **require secrets** — fail closed if any signing/upload/push secret needed for that channel is empty. Missing secrets used to skip Docker Hub, Windows Authenticode, or unstable.golift.io and still go green.
3. **Build: linux / freebsd** (`split` on ubuntu) and **Build: windows** (own job: OIDC + signerd certs stay off the other legs) — `release --clean --split`. Filter with **`GGOOS`**, not `GOOS`. `GOOS` leaks into `go run` before-hooks (man pages, goversioninfo) and they then target the wrong OS. Linux/FreeBSD then run `legacy_gz.sh` (historical `.gz` binary names). FreeBSD then runs `freebsd_txz.sh` (`fpm -t freebsd`).
4. **Build: darwin** (`split-darwin` on macos-latest, skipped on nightly) — import Developer ID + App Store Connect key, same `--split` with `GGOOS=darwin`, staple the DMG.
5. **release N** — download `dist-*` artifacts, import GPG (checksum signatures are created at merge, not split), `continue --merge`. Display name is `release` plus that `REVISION`. This is the only job that pushes Docker / GitHub / AUR / packagecloud / unstable.golift.io. **No Homebrew.**

`REVISION` must be in the goreleaser-action `env:` map (Actions does not automatically forward `GITHUB_ENV` into a later step’s `env:` block). On `--nightly`, `nightly.version_template` already bakes it into `{{ .Version }}` (example `0.9.8-3273`), so man-page hooks use `{{ .Version }}` only — do not append `REVISION` again. The env var is still required: that template *creates* `.Version`, and tagged builds keep `.Version` as the semver while ldflags `Revision` / Darwin `CFBundleVersion` / Windows FileVersion still need the count. Darwin `CFBundleShortVersionString` and Windows FileVersion `x.y.z` use the prefix of `{{ .Version }}`, not `.RawVersion` (that stays on the last tag during `--nightly`).

nFPM `release` is not templated (GoReleaser copies it verbatim). The tagged linux split inserts `release: REVISION` after the `nfpm-release:` marker; `--nightly` leaves it unset. Package files use `{{ replace .ConventionalFileName "~" "-" }}` (GoReleaser `replace` is STRING/OLD/NEW, not sprig). That yields `notifiarr_0.9.8-3273_amd64.deb`, not `_linux_amd64` and not `~.deb`. Do not put `CHANNEL` in the package version. Do not set nFPM `version_metadata`. Tagged FreeBSD packages are `notifiarr-0.9.8_REVISION.amd64.txz`.

## Filename contract (auto-update)

Auto-update **constructs URLs**; it does not scrape a directory listing. Do not rename these assets.

**Unstable** (`pkg/update/unstable.go`): folder `notifiarr`. Payload is a gzipped/zipped **binary**, not a versioned `tar.gz`. Sibling `*.txt` is **JSON**:

```json
{"version":"0.9.8","revision":3273,"size":12345}
```

`version` is the `x.y.z` prefix of GoReleaser `.Version`. `revision` is the integer. `GetUnstable` JSON-decodes those fields and sets `Current` to `version-revision`. A plain-text sidecar (Unpackerr style) 500s the decoder and breaks Synology `jq`.

| Stable URL | Who |
|---|---|
| `https://unstable.golift.io/notifiarr/notifiarr.{GOARCH}.exe.zip` | Windows auto-update (cron + menu) |
| `https://unstable.golift.io/notifiarr/Notifiarr.dmg` | macOS menu download |
| `https://unstable.golift.io/notifiarr/notifiarr.{GOARCH}.{GOOS}.gz` | Linux/FreeBSD (and `userscripts/unstable-syno.sh`) |

**GitHub** (`pkg/update/check.go` `FillUpdate`): first asset whose URL **ends with**

- Windows: `.exe.zip` → `notifiarr.amd64.exe.zip`
- Darwin: `Notifiarr.dmg` (exact). A versioned `Notifiarr_0.9.8_darwin_all.dmg` does **not** match
- FreeBSD: `{amd64,i386,armhf}.txz` → `notifiarr-VERSION.amd64.txz` (etc.)

Linux GitHub matching is already weak (suffix is just `amd64`); in-place auto-update is Windows-only.

**AUR `notifiarr-bin`** (`.github/scripts/aur_publish.sh`) downloads those same GitHub `.linux.gz` names. Keep **binary** AUR: `--split` has no source tarball, and merge Publish is a silent no-op without `aur_sources`.

Historical Linux/FreeBSD `.gz` names are staged after `--split` by `.github/scripts/legacy_gz.sh` and injected into `artifacts.json` as Archive entries (`internal_type: 1`).

## Darwin signing

`.github/scripts/macos_keychain.sh` is **required** on Build: darwin. Missing `MACOS_SIGN_*` / `MACOS_NOTARY_*` fails the job; there is no unsigned-DMG fallback.

`notarize.macos_native.ids` must be the **app bundle** and **DMG** ids (`notifiarr-app`, `notifiarr-dmg`), not the Darwin build id. Those pipes match Extra.ID on the `.app` / `.dmg`. After GoReleaser, `.github/scripts/macos_staple.sh` checks Developer ID on `Notifiarr.app` and staples the DMG (CloudKit can lag a bit after `notarytool` says Accepted). The Darwin `dist/` is packed into one tar before `upload-artifact`; uploading the `.app` tree on macos-latest hangs.

The app executable is `Contents/MacOS/Notifiarr` (`builds.binary: Notifiarr`). Homebrew is unsupported; install from the notarized `Notifiarr.dmg`.

## Merge destinations

- **Docker** — always `ghcr.io/notifiarr/notifiarr` and Hub `docker.io/golift/notifiarr` (`DOCKERHUB_PUBLISH=1`). Empty `DOCKERHUB_PASSWORD` fails the merge job. Platforms: `linux/amd64`, `linux/arm64` (no arm/v7). Three Dockerfiles (runtime COPY of the Pro binary): Alpine, Ubuntu (`-ubuntu`, MegaCli), CUDA (`-cuda`, MegaCli + `nvidia-smi`). The frontend never enters the image. `upload-artifact` zip stores files as `0644`; the merge job `chmod 0755`s `dist/linux/**/notifiarr` and each Dockerfile `COPY --chmod=755` so the image entrypoint is executable.
- **GitHub Release** — tagged `v*` only (`release.disable: "{{ .IsNightly }}"`). macOS is the notarized `Notifiarr.dmg`. Windows assets are `notifiarr.amd64.exe.zip`. FreeBSD assets are pkgng `notifiarr-<version>.{amd64,i386,armhf}.txz` plus `.freebsd.gz`. Linux assets are `notifiarr.{amd64,386,arm,arm64}.linux.gz` plus nFPM deb/rpm/zst.
- **AUR** — tagged `CHANNEL=release` only, after packagecloud. `.github/scripts/aur_publish.sh` hashes `dist/linux/notifiarr.*.linux.gz` from `legacy_gz.sh` (CI fails if those files are missing). Binary package `notifiarr-bin`. Do not add GoReleaser `aur_sources` (`--split` fatals `no linux archives found`; merge Publish is a silent no-op). Skip nightly/unstable. nFPM `.pkg.tar.zst` on the GitHub Release is a separate binary package.
- **packagecloud** — `golift/pkgs` vs `golift/unstable`. Skip when `CHANNEL=nightly`.
- **unstable.golift.io** — only `CHANNEL=unstable` (tagged releases do **not** clobber testers). Script: `.github/scripts/unstable_upload.sh`. Upload overwrites by name. Empty `UNSTABLE_UPLOAD_KEY` fails in GitHub Actions.
- **Homebrew** — none. Do not require `HOMEBREW_TAP_GITHUB_TOKEN`.

Linux nFPM names are conventional (`notifiarr_0.9.8-3273_amd64.deb`, `notifiarr-0.9.8-3273.x86_64.rpm`). Arches: amd64, arm64, i386, armv7 (one `armhf`). Darwin min macOS 13. `CFBundleShortVersionString` and Windows FileVersion both take the `x.y.z` prefix of `{{ .Version }}` (not `.RawVersion`); Windows FileVersion’s fourth number is `REVISION`. Windows is `-H=windowsgui` without `netgo`. Empty `CODESIGN_URL` fails in GitHub Actions (local snapshots still skip).

## Secrets

Set on the `Notifiarr/notifiarr` repo (or org, granted to this public repo). `.github/scripts/require_secrets.sh` runs before any build job and **fails the workflow** if a secret required for that `CHANNEL` is empty. Nightly does not require Apple / AUR / packagecloud / unstable-upload secrets (those destinations are skipped). Confirm `GORELEASER_PRO_KEY` is granted to this public repo.

| Secret | Used by |
|---|---|
| `GORELEASER_PRO_KEY` | every goreleaser-action (env `GORELEASER_KEY`) |
| `GPG_SIGNING_KEY` | Linux nFPM signatures |
| `MACOS_SIGN_P12`, `MACOS_SIGN_PASSWORD` | Developer ID `.p12` (PEM or long-line base64) |
| `MACOS_NOTARY_KEY`, `MACOS_NOTARY_KEY_ID`, `MACOS_NOTARY_ISSUER_ID` | App Store Connect `.p8` |
| `CODESIGN_URL`, `CODESIGN_CLIENT_CERT`, `CODESIGN_CLIENT_KEY` | Windows Authenticode (OIDC + mTLS) |
| `DOCKERHUB_PASSWORD` | Hub login (required) |
| `PACKAGECLOUD_TOKEN` | `golift/pkgs` / `golift/unstable` |
| `AUR_DEPLOY_KEY` | AUR `notifiarr-bin` (release channel) |
| `UNSTABLE_UPLOAD_KEY` | unstable.golift.io (unstable channel) |

`GITHUB_TOKEN` is the default Actions token (GHCR + GitHub Releases). There is no `HOMEBREW_TAP_GITHUB_TOKEN`.

## Action pins

`release.yml` pins `owner/repo@<commit-sha> # vX.Y.Z`. Floating major tags (`@v4`) are not used.

The Alpine Docker base image is pinned as `alpine:<tag>@sha256:<digest>`. Renovate keeps Action and Dockerfile digest pins current (`helpers:pinGitHubActionDigestsToSemver`; Dockerfile `pinDigests`).

Renovate automerges Go and Docker non-major updates, and GitHub Actions updates including majors, after a 7-day release age when checks pass.
