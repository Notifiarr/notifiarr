# Contributing to Notifiarr

Notifiarr is a Go and Svelte client for Notifiarr.com. Contributions should follow the existing client, integrations, service checks, web UI, and platform packaging.

[AGENTS.md](AGENTS.md) is the command reference: build targets, the CI test shortcut, generated files, logging, and lint. Use it when you need the exact commands.

## Development setup

1. Install the Go version required by `go.mod`.
2. Install Node 24 and npm. Frontend CI uses Node 24.
3. From the repository root, run `make generate` before building. That refreshes API types, OpenAPI specs, locales, and the frontend bundle.

The example configuration is `examples/notifiarr.conf.example`. `examples/MANUAL.md` documents `notifiarr -c <config file>` and the other flags.

`make dev` generates assets, builds a race-enabled binary, and starts the client with `DN_ENCODE_CONFIG_FILE=false`, so a new config file stays plain text. `make notifiarr` builds without starting the client. Without that environment variable, `pkg/configfile` bzip2-encodes a config file when it creates one.

## Checks

For a Go-only change, match pull-request CI:

```sh
mkdir -p frontend/dist
echo "Fake frontend build." > frontend/dist/index.html
go test ./pkg/...
```

`make test` is the full local run. It deletes `./notifiarr`, generates assets (including the npm build), lints, then runs `go test -race -covermode=atomic ./...`.

Format Go with `golangci-lint fmt` (golangci-lint v2). That applies `gci`, `gofmt`, `gofumpt`, and `goimports`. `make lint` also runs codespell and golangci-lint for Linux, Darwin, FreeBSD, and Windows.

Frontend checks, from `frontend` after `npm ci`:

```sh
npm run check
npm run lint
npm run format:check
npm test
npm run build
```

The workflows are `.github/workflows/codetests.yml` and `.github/workflows/frontend.yml`.

## Code style

Keep changes inside the existing `pkg/` boundaries. Wrap errors with `%w` when you add context, and use `errors.Is` for sentinel errors. Log with `mnd.Log` and a request id. Keep platform-specific code in its existing file.

Change generator inputs instead of editing generated output. Commit the regenerated files: `frontend/locales.go`, `frontend/public/api_openapi.json`, `frontend/public/ui_openapi.json`, and `frontend/src/api/notifiarrConfig.ts`.

## Commits and pull requests

Use a short imperative commit subject. Do not append a pull-request number. GitHub adds that when the pull request is squash-merged.

Issue titles created from the GitHub forms use the prefixes `Bug - ` and `FR - `.

[Discuss a large change on Discord](https://notifiarr.com/discord) first. Open the pull request against [Notifiarr/notifiarr](https://github.com/Notifiarr/notifiarr) `main`, and name the commands you ran.
