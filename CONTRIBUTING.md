# Contributing

Issues and pull requests are welcome. For a substantial change, open an issue
first.

## Setup

Go (the version in `go.mod`), Helm and [gitleaks](https://github.com/gitleaks/gitleaks).
`make tools` installs the pinned Go tools (golangci-lint, govulncheck,
gocover-cobertura). [prek](https://github.com/j178/prek) runs the git hooks:
`prek install`.

## Checks

`make check` is the definition of green: lint (Go, Helm chart, workflows),
tests with the race detector, coverage of at least 80 %, vulnerability, license
and secret scans, the build and `fleetlint check`. CI runs the same target.
`make help` lists all targets.

The cert-manager conformance suite needs envtest binaries and a real KAS
account and writes real DNS records — it is opt-in, see
[testdata/allinkl/README.md](testdata/allinkl/README.md).

## Pull requests

- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)
  (`feat:`, `fix:`, `chore(deps):` …); the commit-msg hook checks them.
- New behaviour comes with a test against the kasapi fake server
  (`kasapitest`) and, for Kubernetes access, the client-go fake.
