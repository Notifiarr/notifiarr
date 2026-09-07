# Contributing to Notifiarr

Notifiarr is a Go and Svelte client for Notifiarr.com. Contributions should fit the existing client, integrations, service checks, web UI, and platform packaging structure.

## Development setup

1. Install the latest Go, as required by `go.mod`.
2. Install the frontend tools required by `frontend/package.json` and ensure `npm` is available. `make generate` runs `npm install` and `npm run build` through `frontend/generate.sh`.
3. Run `make generate` from the repository root before building or testing generated API and frontend output.

There is no `deps` target in the repository's `Makefile`, so `make deps` is not an available setup command. Go dependencies are resolved by the Go toolchain; API generation runs `go mod vendor` temporarily through `frontend/src/api/generate.sh` when needed.

## Build, test, and lint

| Command | Result |
| --- | --- |
| `make generate` | Generate API definitions and frontend assets. |
| `make notifiarr` | Build the `notifiarr` executable. |
| `make dev` | Build with the Go race detector and run the client. |
| `go test ./...` | Run all Go package tests. |
| `make test` | Clean, generate, lint, then run race-enabled Go tests with coverage. |
| `golangci-lint run` | Lint the current platform. |
| `make lint` | Generate, run `codespell`, and lint Linux, Darwin, FreeBSD, and Windows configurations. |

The Go linter configuration is `.golangci.yml`. It enables `gci`, `gofmt`, `gofumpt`, and `goimports` formatters. Format Go changes with `gofmt` and `goimports`, then check the result with the configured lint commands.

Frontend checks run from `frontend`:

```sh
npm run check
npm run lint
npm run format:check
npm test
```

The same checks are defined in `.github/workflows/frontend.yml`. The Go test and lint workflow is `.github/workflows/codetests.yml`.

## Local execution

Use `make dev` to generate assets, build a race-enabled binary, and start the client. For a non-running build, use `make notifiarr`. The example configuration is `examples/notifiarr.conf.example`; `examples/MANUAL.md` documents `notifiarr -c <config file>` and the available command-line options.

## Code style

Follow the existing Go package boundaries under `pkg/`. Handle errors explicitly, wrap errors with `%w` when adding context, and use `errors.Is` with package-defined sentinel errors. Preserve the repository's standard-library logging patterns. Keep platform-specific implementations in their existing platform files and update generator inputs rather than editing generated output directly.

## Commits and pull requests

Recent history uses short, imperative commit subjects. Dependency updates and focused fixes commonly include the related PR number, such as `Update dependency ... (#1345)`. Keep commits focused and describe behavior changes clearly.

The repository has no checked-in `CONTRIBUTING.md` under `.github/` and no pull-request template. The root `README.md` asks contributors to discuss proposals in [Notifiarr Discord](https://notifiarr.com/discord). After that discussion, create a branch, make the focused change, run the relevant checks, and open a pull request to [Notifiarr/notifiarr](https://github.com/Notifiarr/notifiarr). The repository's pull-request workflows target `main`, `unstable`, and `development`; select the branch that matches the work and state the verification commands in the PR description.
