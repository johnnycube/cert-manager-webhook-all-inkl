# Changelog

All notable changes to this project are documented here.
Format: Keep a Changelog. Versioning: Semantic Versioning. Entries up to 0.4.2
are taken from the GitHub release notes.

## [Unreleased]

## [0.4.4] - 2026-10-21

### Security

- Built with Go 1.26.10, which distrusts compromised intermediate certificates
  in crypto/x509 (GO-2026-6809, CVE-2026-111943).
- cel-go 0.30.0 (GO-2026-6094) and etcd client/pkg 3.6.14 (GO-2026-6107).
  Both are transitive dependencies of the Kubernetes libraries; the webhook
  does not call the affected code.

### Changed

- kasapi 0.3.4 and cert-manager 1.21.2. No changes to the webhook code.

## [0.4.3] - 2026-10-14

### Changed

- kasapi 0.3.3.
- `Present` and `CleanUp` share one path that reads the config, the
  credentials and the record name. Errors carry one `allinkl:` prefix instead
  of two. `github.com/pkg/errors` is no longer a dependency.

### Security

- The image is signed with cosign (keyless) by digest and carries an SBOM and
  a SLSA provenance attestation. The Helm chart is signed by digest.
- The Dockerfile pins the Go build image by digest.

## [0.4.2] - 2026-10-14

### Security

- Built with Go 1.26.9 (standard library fixes for net/http, crypto/tls and
  html/template) and golang.org/x/net 0.60.0 (HTTP/2 server fixes).

### Changed

- kasapi 0.3.2, which includes the kasapi 0.3.1 bump that missed 0.4.1.
- docker/dockerfile 1.28. No changes to the webhook code.

## [0.4.1] - 2026-09-29

### Changed

- Dependency and base image refresh: grpc 1.83.2, cel-go 0.29.0,
  OpenTelemetry 1.45.0 and updated distroless base image digests. No changes
  to the webhook code.

## [0.4.0] - 2026-07-20

### Changed

- The KAS API calls go through the [kasapi](https://github.com/johnnycube/kasapi)
  module instead of code copied from lego.
- Dependency updates.

## [0.3.0] - 2025-10-24

### Changed

- No stateful KAS client management: cleanup uses the request cert-manager
  sends instead of a stored reference.

## [0.2.0] - 2025-10-20

### Changed

- Dependency updates: actions/checkout v5, docker/dockerfile 1.19,
  cert-manager 1.19.1; dependencies pinned.

## [0.1.0] - 2025-10-18

### Added

- First working release: normal domains, wildcards and sub-subdomains.

[Unreleased]: https://github.com/johnnycube/cert-manager-webhook-all-inkl/compare/v0.4.4...HEAD
[0.4.4]: https://github.com/johnnycube/cert-manager-webhook-all-inkl/compare/v0.4.3...v0.4.4
[0.4.3]: https://github.com/johnnycube/cert-manager-webhook-all-inkl/compare/v0.4.2...v0.4.3
[0.4.2]: https://github.com/johnnycube/cert-manager-webhook-all-inkl/compare/v0.4.1...v0.4.2
[0.4.1]: https://github.com/johnnycube/cert-manager-webhook-all-inkl/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/johnnycube/cert-manager-webhook-all-inkl/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/johnnycube/cert-manager-webhook-all-inkl/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/johnnycube/cert-manager-webhook-all-inkl/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/johnnycube/cert-manager-webhook-all-inkl/releases/tag/v0.1.0
