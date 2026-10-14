# Task-runner contract (https://fleetlint.org/baseline/#task-runner-contract). `make check` is the definition of green.
SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.ONESHELL:
.DEFAULT_GOAL := check-fast
MAIN_BRANCH ?= main
COVER_MIN   ?= 80
PKGS        := ./...
BIN         := bin/webhook-allinkl
CHART       := deploy/helm
IMAGE       ?= ghcr.io/johnnycube/cert-manager-webhook-all-inkl:dev
ENGINE      ?= $(shell command -v podman >/dev/null && echo podman || echo docker)
LICENSES    ?= MIT,Apache-2.0,BSD-2-Clause,BSD-3-Clause,ISC
# Each pin carries a `# renovate:` comment so Renovate bumps it (customManagers:makefileVersions).
# renovate: datasource=go depName=github.com/golangci/golangci-lint/v2
GOLANGCI_VERSION    ?= v2.14.0
# renovate: datasource=go depName=golang.org/x/vuln
GOVULNCHECK_VERSION ?= v1.9.0
# renovate: datasource=go depName=github.com/boumenot/gocover-cobertura
GOCOVER_VERSION     ?= v1.5.0
# renovate: datasource=go depName=github.com/rhysd/actionlint
ACTIONLINT_VERSION  ?= v1.7.12
# renovate: datasource=go depName=github.com/google/go-licenses/v2
GO_LICENSES_VERSION ?= v2.0.1
ACTIONLINT  := go run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION)
# Tools are built with this module's toolchain: golangci-lint refuses code that targets a newer Go than it was built with.
TOOLCHAIN   := $(shell go env GOVERSION)
CHART_SEMVER := $(shell awk '/^version:/{print $$2}' $(CHART)/Chart.yaml)

.PHONY: help tools fmt lint test cover audit bench build image dist check-fast check release

help: ## list targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | awk -F':.*##' '{printf "  %-12s %s\n", $$1, $$2}'

tools: ## install the pinned Go tools the gates use (gitleaks, helm and diff-cover come from their own installers; actionlint runs via go run)
	GOTOOLCHAIN=$(TOOLCHAIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_VERSION)
	GOTOOLCHAIN=$(TOOLCHAIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	GOTOOLCHAIN=$(TOOLCHAIN) go install github.com/boumenot/gocover-cobertura@$(GOCOVER_VERSION)
	GOTOOLCHAIN=$(TOOLCHAIN) go install github.com/google/go-licenses/v2@$(GO_LICENSES_VERSION)

fmt: ## format in place
	golangci-lint fmt

lint: ## format check, linters (also over the conformance test), vet, chart and workflow lint
	golangci-lint fmt --diff
	golangci-lint run
	golangci-lint run --build-tags conformance
	go vet $(PKGS)
	helm lint $(CHART)
	$(ACTIONLINT)

test: ## unit tests with the race detector, shuffled (the conformance suite is opt-in: go test -tags conformance)
	go test -race -shuffle=on $(PKGS)

cover: ## total coverage >= COVER_MIN, plus changed lines >= COVER_MIN when diff-cover is installed
	mkdir -p coverage
	go test -race -coverprofile=coverage/cover.out $(PKGS)
	total=$$(go tool cover -func=coverage/cover.out | awk '/^total:/{sub("%","",$$3); print $$3}')
	echo "total coverage: $$total% (minimum $(COVER_MIN)%)"
	awk -v t="$$total" -v m="$(COVER_MIN)" 'BEGIN{exit !(t >= m)}' || { echo "coverage below $(COVER_MIN)%"; exit 1; }
	if command -v gocover-cobertura >/dev/null && command -v diff-cover >/dev/null; then
	  gocover-cobertura < coverage/cover.out > coverage/cobertura.xml
	  diff-cover coverage/cobertura.xml --compare-branch=origin/$(MAIN_BRANCH) --fail-under=$(COVER_MIN)
	else echo "diff-cover/gocover-cobertura not installed: changed-line gate skipped"; fi

audit: ## vulnerabilities, licenses, tidy module graph, secrets
	govulncheck $(PKGS)
	go-licenses check $(PKGS) --allowed_licenses=$(LICENSES)
	go mod tidy -diff
	gitleaks git --no-banner --redact .

bench: ## benchmarks (not part of check)
	go test -run='^$$' -bench=. -benchmem $(PKGS)

build: ## build the webhook binary
	CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o $(BIN) .

image: ## build the container image locally (podman or docker)
	$(ENGINE) build -t $(IMAGE) .

dist: ## package the Helm chart into dist/ (the image is built and pushed by docker.yaml)
	helm package $(CHART) --destination dist

check-fast: ## before every commit (< 30 s)
	golangci-lint fmt --diff
	golangci-lint run --new-from-rev=origin/$(MAIN_BRANCH) || golangci-lint run
	go test -short $(PKGS)

check: lint test cover audit build ## definition of green (< 10 min)
	fleetlint check --fail-on error

release: ## on a clean main whose Chart.yaml version is new: check, then sign tag v<chart version>; pushing the tag publishes image and chart
	@test -z "$$(git status --porcelain)" || { echo "dirty tree"; exit 1; }
	@test "$$(git branch --show-current)" = "$(MAIN_BRANCH)" || { echo "not on $(MAIN_BRANCH)"; exit 1; }
	@! git rev-parse -q --verify "refs/tags/v$(CHART_SEMVER)" >/dev/null || { echo "v$(CHART_SEMVER) exists: bump $(CHART)/Chart.yaml first"; exit 1; }
	@grep -q "^## \[$(CHART_SEMVER)\]" CHANGELOG.md || { echo "CHANGELOG.md has no section for $(CHART_SEMVER)"; exit 1; }
	$(MAKE) check
	git tag -s "v$(CHART_SEMVER)" -m "v$(CHART_SEMVER)"
	@echo "tagged v$(CHART_SEMVER) — push with: git push origin v$(CHART_SEMVER)"
