# cert-manager-webhook-all-inkl
A cert-manager webhook for all-inkl.com hosting

## How to use

### Deploy with Helm

An OCI helm chart is published as part of the ghcr packages: https://github.com/johnnycube/cert-manager-webhook-all-inkl/pkgs/container/allinkl-webhook

To install use your favourite helm tool:

```bash
helm install allinkl-webhook oci://ghcr.io/johnnycube/allinkl-webhook-helm --version 0.1.0
```

### Create (Cluster-) Issuer and Certificate

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: excample-com-tls
  namespace: cert-manager
spec:
  secretName: demo-tls
  dnsNames:
    - example.com
  issuerRef:
    name: allinkl-issuer # or ClusterIssuer with kind: ClusterIssuer
    kind: Issuer
```

The referenced credentials have to be deployed in a secret readable by the Service Acount used by the webhook. `cert-manager` is the default namespace

```yaml
---
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: allinkl-issuer
  namespace: cert-manager
spec:
  acme:
    email: test@example.com # Change to your letsencrypt email
    server: https://acme-staging-v02.api.letsencrypt.org/directory # Change for productive
    privateKeySecretRef:
      name: allinkl-webhook-account-key
    solvers:
      - dns01:
          webhook:
            groupName: acme.johanneskueber.com
            solverName: allinkl
            config:
              usernameKeySecretRef:
                name: allinkl-credentials
                key: username
              passwordKeySecretRef:
                name: allinkl-credentials
                key: password
```

### Use with Ingress / Gateway API

## Road to 1.0

* [] Add tests
* [] Ensure that the cert-manager testsuite is running
* [] Proper documentation / README.md
* [] RBAC for ServiceAccount in Helm

## Licensing & Attributions

This project includes code derived from [go-acme/lego] (MIT).
See `THIRD_PARTY_NOTICES.md`.