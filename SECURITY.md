# Security policy

Report suspected vulnerabilities privately through GitHub's "Report a
vulnerability" button on the Security tab, or by email to
`security@johanneskueber.com`. Do not open public issues for security reports.

Reports are acknowledged within a few days. The supported version is the latest
release.

The webhook holds KAS credentials from Kubernetes Secrets and writes TXT records
in the account's zones. Vulnerabilities in the KAS service itself go to
all-inkl.com directly; vulnerabilities in cert-manager go to the cert-manager
project.

Images and charts are signed with cosign (keyless, GitHub OIDC). Verify an
image before deploying it:

```sh
cosign verify ghcr.io/johnnycube/cert-manager-webhook-all-inkl:<tag> \
  --certificate-identity-regexp '^https://github.com/johnnycube/cert-manager-webhook-all-inkl/' \
  --certificate-oidc-issuer https://token.actions.githubusercontent.com
```
