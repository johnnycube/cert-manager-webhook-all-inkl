---
name: Bug report
about: The webhook does not create or remove a record as expected
labels: bug
---

**Versions**: webhook image/chart, cert-manager, Kubernetes

**What happened**: the challenge's state and the webhook log lines around it
(`kubectl logs deploy/allinkl-webhook`; remove credentials)

**Expected**:

**Steps to reproduce**: Issuer and Certificate (credentials removed)
