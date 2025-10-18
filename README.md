# cert-manager-webhook-all-inkl
A {cert-manager](https://cert-manager.io/) [webhook](https://cert-manager.io/docs/concepts/webhook/) for [all-inkl.com](https://all-inkl.com/) hosting

The library is based on the [official example](https://github.com/cert-manager/webhook-example) from cert-manager.

## How to use

The webhook is meant to be run on a kubernetes cluster. It uses kubernetes apis to register a webhook and will not work without.

### Deploy with Helm

An OCI helm chart is published as part of the [ghcr packages](https://github.com/johnnycube/cert-manager-webhook-all-inkl/pkgs/container/allinkl-webhook)

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
    name: allinkl-issuer # or Issuer with kind: Issuer
    kind: ClusterIssuer
```

The referenced credentials have to be deployed in a secret readable by the Service Acount used by the webhook. `cert-manager` is the default namespace. This is also the namespace used for ClusterIssuers.

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

Username and password are the credentials for your [KAS](https://kas.all-inkl.com/).


### Use with Ingress / Gateway API

Cert manager can automate the creation of certificates when used in conjuntion with the [Gateway API](https://gateway-api.sigs.k8s.io/). To use this feature your gateway needs to be annotated with a cert-manager specific annotation. Also don't forget to activate the [gateway API support](https://cert-manager.io/docs/usage/gateway/) of cert-manager - deactivated by default on older versions.

```yaml
---
# Gateway for *.example.com
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: gw-example
  namespace: default
  annotations:
    cert-manager.io/cluster-issuer: allinkl-issuer # or Issuer with annotation: cert-manager.io/issuer
spec:
  gatewayClassName: traefik # adept to your gwc
  listeners:
    - name: https
      protocol: HTTPS
      port: 443
      hostname: "*.example.com"
      tls:
        mode: Terminate
        certificateRefs:
          - kind: Secret
            name: wildcard-example-com-tls
            namespace: default
      allowedRoutes:
        namespaces:
          from: All
```

## Road to 1.0

* [] Add (unit) tests
* [] Ensure that the cert-manager testsuite is running
* [] Proper documentation / README.md
* [] Describe solution RBAC for ServiceAccount in Helm

## Attributions

This project includes code derived from [go-acme/lego](https://github.com/go-acme/lego) (MIT).<br />
This project includes code derived from [cert-manager/webhook-example](https://github.com/cert-manager/webhook-example) (Apache v2).<br />
See `THIRD_PARTY_NOTICES.md`.

## Contributions

If you've found an error in this sample, please file an issue.

Patches are encouraged and may be submitted by forking this project and
submitting a pull request. Since this project is still in its very early stages,
if your change is substantial, please raise an issue first to discuss it.

## License

```
MIT License

Copyright (c) 2025 Johannes Küber

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```
