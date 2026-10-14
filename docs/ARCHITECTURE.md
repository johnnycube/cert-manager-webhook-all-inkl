# Architecture

The webhook is a cert-manager DNS-01 solver for zones hosted at all-inkl.com.
cert-manager calls it to create the `_acme-challenge` TXT record for a
certificate order and to remove it afterwards; the webhook turns those calls
into KAS API requests.

## Components

- **`main.go`** — starts the webhook server from cert-manager's
  `pkg/acme/webhook/cmd`. That library serves the extension API on :443 with
  TLS, registers it under `GROUP_NAME` and dispatches challenges to the solver.
- **`pkg/solver`** — the solver. `solver.go` implements `Present`, `CleanUp`
  and `Initialize`; `config.go` decodes the per-Issuer config; `client.go`
  creates and deletes TXT records through
  [kasapi](https://github.com/johnnycube/kasapi).
- **`deploy/helm`** — the chart: Deployment, Service, the `APIService` that
  makes the Kubernetes API server proxy `v1alpha1.<groupName>` to the webhook,
  the PKI for that connection (cert-manager `Issuer`/`Certificate`) and RBAC.

## Data flow

1. A `Certificate` with a DNS-01 webhook solver makes cert-manager create a
   `ChallengeRequest` and send it through the Kubernetes API server to the
   webhook's `APIService`.
2. `Present` decodes the solver config from the request, reads the KAS login
   and password from the referenced Secret in the request's namespace and
   computes the record name relative to the resolved zone.
3. The client opens a KAS session and sends `add_dns_settings` for a TXT record
   with the challenge key.
4. After validation cert-manager calls `CleanUp`, which repeats step 2, lists
   the zone with `get_dns_settings` and sends `delete_dns_settings` for each TXT
   record whose name *and* value match the challenge.

## Boundaries

- The webhook is stateless — every call carries what it needs. Nothing is
  cached between `Present` and `CleanUp`, so a restart between them is
  harmless.
- Credentials live only in Kubernetes Secrets. The webhook reads them per call
  with its ServiceAccount and never logs them; the RBAC for that Secret is the
  operator's to grant (see the README).
- KAS session handling and flood-protection delays are kasapi's job. The
  webhook does not retry on its own; cert-manager retries a failed challenge.

## Decisions

- **Delete by name and value.** Concurrent orders for the same name (a wildcard
  and its apex) share `_acme-challenge`. Matching the value removes only this
  challenge's record.
- **`kasapi` instead of a local KAS client.** Up to 0.3 the webhook carried KAS
  code copied from lego. The shared library is tested against a fake KAS server
  and is used by the Terraform provider as well.
- **Plain session auth.** Many KAS accounts have SHA-1 session auth disabled
  (`kas_auth_type_disabled`); the client uses `AuthPlain` over HTTPS.
