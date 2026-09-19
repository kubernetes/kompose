# AGENTS.md

Kompose converts `docker-compose` (Compose Spec) files into Kubernetes or OpenShift manifests. CLI entrypoint is `main.go` -> `cmd/` (cobra).

## Build & quick checks

- Build: `make bin` (CGO disabled, injects `pkg/version.GITCOMMIT`), or `go build -o kompose main.go`.
- Run a single Go package's tests: `go test -short ./pkg/transformer/kubernetes/...` (always use `-short`; `script/test/cmd/cmd_test.go` is skipped when short).
- Fast validation: `make validate` = gofmt check + `go vet ./pkg/...`. CI's lint workflow only runs `go vet ./pkg/...`.
- `make golangci-lint` installs golangci-lint **v1.32.2** (very old, via a `go-get-tool` helper into `./bin`); it may not build under the modern Go required by `go.mod` (go 1.26.8). Prefer `go vet` + gofmt locally.

## Full test suite

`make test` is heavyweight and needs prerequisites:
1. Installs `goveralls`, `gover`, `gox` via `test-dep`.
2. Runs unit tests per-package with `go test -short -race -cover`, aggregating into `gover.coverprofile` at repo root (**this file is committed; don't commit regenerated coverage by mistake**).
3. `go install` puts the `kompose` binary on `$PATH`.
4. Runs CLI/e2e tests (`script/test/cmd/tests.sh`), which require:
   - `dyff` on PATH — install with `go install github.com/homeport/dyff/cmd/dyff@v1.5.8` (CI installs it before `make test`).
   - Docker or podman running, else the `--build local` test cases are auto-skipped.
5. `make cross` cross-compiles for linux/arm/arm64, windows/amd64, darwin/amd64+arm64.

## CLI e2e golden tests (`script/test/cmd/tests.sh`)

- Each fixture lives in `script/test/fixtures/<name>/` with `compose.yaml`, an `output-k8s.yaml` and (for openshift) `output-os.yaml`. Output is compared with `dyff between --ignore-order-changes`.
- Changing conversion output (or adding a flag) usually means updating BOTH golden files and adding a `convert::expect_success[_and_warning]` case in `tests.sh`.
- Most tests run for both providers: default `kubernetes` and `--provider=openshift`.
- Modern compose example that exercises most conversion paths: `examples/compose.yaml`.

## Architecture / data flow

- `pkg/kobject` — the internal intermediate model (`ServiceConfig`, `ConvertOptions`); source-package-neutral.
- `pkg/loader/compose` — parses compose (uses upstream `compose-go/v2` types directly, `kompose`'s own Compose wrapper) -> `kobject`.
- `pkg/transformer/{kubernetes,openshift}` — `kobject` -> manifests. `pkg/app` orchestrates loader + transformer.
- `client/` — programmatic library API; `cmd/` is the thin cobra CLI over it.
- Go unit tests in `pkg/transformer/...` and `pkg/loader/compose/...` construct `kobject.ServiceConfig` inline (no fixtures).

## Repo conventions & gotchas

- Commit messages must follow [Conventional Commits](https://www.conventionalcommits.org/) (enforced by `.pre-commit-config.yaml` via commitlint). Pre-commit hooks also run gofmt/goimports/golangci-lint/unit tests.
- Kubernetes API version is locked to what OpenShift's fork uses — do not bump `k8s.io/*` independently (see `docs/development.md`); `go.mod` also carries a `replace` for `openshift/api`.
- `docs/` is the Jekyll-backed kompose.io website source, not just project docs; `index.md` at repo root is site content, not the README.
- `.gitmodules` is empty; `gover.coverprofile` is intentionally committed.
