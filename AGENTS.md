# Notifiarr agent guide

## Project

Notifiarr is the unified client for Notifiarr.com. It is written in Go and Svelte and runs on most operating systems. The client runs a web server, monitors local services, sends system and Plex data to Notifiarr, and supports media requests from Discord. See `README.md` and `examples/MANUAL.md` for the user-facing scope and command-line options.

## Build and run

The module requires the Go version in `go.mod`. Frontend CI uses Node 24. Run commands from the repository root. The Makefile loads version metadata from `settings.sh`.

| Command | Purpose |
| --- | --- |
| `make generate` | Generate API types, OpenAPI specs, locales, and the frontend bundle. |
| `make notifiarr` | Generate, then build `./notifiarr` with `-tags osusergo,netgo`. |
| `make dev` | Generate, build a race-enabled `./notifiarr`, and run it with `DN_ENCODE_CONFIG_FILE=false`. |
| `make clean` | Remove `./notifiarr`, generated man pages, and release artifacts. |

`make generate` writes a placeholder `frontend/dist/index.html`, runs `go generate ./frontend/src/api`, then `go generate ./frontend`. `frontend/generate.sh` runs `npm install` and `npm run build`, then writes `frontend/locales.go` and the OpenAPI JSON. Do not commit `frontend/dist`. It exists so `//go:embed dist` in `frontend/frontend.go` can compile. `pkg/client` imports `frontend`, so Go tests need that directory too.

`make dev` is the usual local run. The example configuration is `examples/notifiarr.conf.example`. `examples/MANUAL.md` documents `-c` / `--config` and the other flags.

Leave `DN_ENCODE_CONFIG_FILE=false` set when a new config file should stay plain text. Otherwise `pkg/configfile` bzip2-encodes the file when it creates one.

Release tags live in `.goreleaser.yaml`. Linux and Darwin use `osusergo,netgo`. FreeBSD adds `nodbus`. Windows uses `osusergo` only, plus `-H=windowsgui`. Darwin release builds need CGO for the systray Cocoa backend. `.github/workflows/README.md` describes how those builds are published. `unstable` is a manual publish branch (`git push origin main:unstable`), not a feature branch.

## Tests and lint

Go-only changes can match pull-request CI without an npm build:

```sh
mkdir -p frontend/dist
echo "Fake frontend build." > frontend/dist/index.html
go test ./pkg/...
```

`.github/workflows/codetests.yml` runs that same snippet on Linux, macOS, and Windows. The test step does not pass `-race`.

`make test` is the heavy local path. It runs `make clean` (this deletes `./notifiarr`), `make generate` (including the npm build), `make lint`, then `go test -race -covermode=atomic ./...`.

`make lint` runs `make generate`, codespell, and golangci-lint for Linux, Darwin, FreeBSD (`--build-tags nodbus`), and Windows. CI lints the same four `GOOS` values and also passes `--build-tags nodbus` for the Darwin job. The config is `.golangci.yml` (golangci-lint v2; CI pins v2.13). Format Go with `golangci-lint fmt` so `gci`, `gofmt`, `gofumpt`, and `goimports` all run.

Frontend checks run from `frontend` on Node 24. CI uses `npm ci`, then:

```sh
npm run check
npm run lint
npm run format:check
npm test
npm run build
```

`npm run check` is svelte-check. `npm test` runs Vitest once. `.github/workflows/frontend.yml` on `main` runs these steps and fails the job when Prettier or `npm run build` fails.

## Generated files

Change the inputs, regenerate, and commit the outputs. Do not hand-edit these files.

| Output | Input |
| --- | --- |
| `frontend/locales.go` | `frontend/src/includes/locale/*.json`, via `frontend/generate.sh` |
| `frontend/public/api_openapi.json` and `frontend/public/ui_openapi.json` | Swagger comments, via `frontend/generate.sh`. UI tags are `System`, `Integrations`, and `Files`. Every other tag is the API spec. |
| `frontend/src/api/notifiarrConfig.ts` | Go config types, via `go generate ./frontend/src/api`. That script runs `go mod vendor` so it can read dependency docs, then deletes `vendor/`. |

## Layout

- `main.go` starts `pkg/client`.
- `pkg/client` owns startup, the web server, handlers, the GUI profile, and tray commands. It serves the Svelte UI through `frontend.IndexHandler`.
- `pkg/apps` configures Radarr, Lidarr, Sonarr, Readarr, Prowlarr, qBittorrent, and Transmission. Plex, SABnzbd, and Tautulli live in `pkg/apps/apppkg`. Starr HTTP clients come from `golift.io/starr`.
- `pkg/services` implements service and network checks. `pkg/checkapp` contains application checks.
- `pkg/configfile` reads and writes the config file, including the optional bzip2 encoding. `pkg/ui` is platform-specific UI behavior.
- `pkg/snapshot` collects system, disk, process, RAID, SMART, NVIDIA, MySQL, and user information. `pkg/mnd` holds constants, the shared `Log` logger, Prometheus support, and shared types.
- `pkg/triggers` holds autoupdate, backups, cfsync, commands, dashboard, emptytrash, endpoints, fileupload, filewatch, gaps, mdblist, plexcron, snapcron, and starrqueue. `pkg/website` talks to Notifiarr.com. `pkg/website/clientinfo` builds the check-in payload.
- `pkg/bindata` embeds extra files. `pkg/cooldown`, `pkg/logs`, `pkg/private`, and `pkg/update` are supporting packages.
- `frontend` is the Svelte UI, API generation, locales, tests, and build scripts. `examples`, `init`, and `userscripts` hold configuration examples, packaging and service files, and install scripts.

## Code conventions

Stay inside the existing `pkg/` boundaries. Wrap errors with `fmt.Errorf("...: %w", err)` when adding context, and use `errors.Is` for sentinel errors the package already exposes. Log through `mnd.Log` (`Printf`, `Errorf`, `Debugf`, `Trace`) and pass a request id, matching the surrounding call.

Run `golangci-lint fmt` before lint. `err113` is disabled for `pkg/**/*_test.go`, so tests may build errors inline. Keep platform-specific code in the existing `*_linux.go`, `*_darwin.go`, `*_windows.go`, and `*_freebsd.go` files.

## Commits and pull requests

Use a short imperative subject, such as `Fix Sonarr queue retry` or `docs: describe the CI test shortcut`. Leave off a `(#1234)` suffix. GitHub adds the pull-request number when it squash-merges.

Open pull requests against `main`. The test workflow also lists `unstable` and `development`. Feature work still targets `main`. Discuss a large change in the Discord linked from `README.md`. In the pull request, describe the behavior change and the commands you ran.
