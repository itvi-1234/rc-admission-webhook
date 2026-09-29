# rc-admission-webhook

Validating admission webhook for `RuntimeConditionsProfile` custom resources.

## Prerequisites

- The `RuntimeConditionsProfile` CRD (from `runtime-conditions-crd`) installed in the cluster.
- cert-manager. Admission webhooks are TLS-only, and cert-manager issues and
  rotates the webhook's serving certificate. If your cluster doesn't have it
  yet, set `certManager.installDependency=true` to pull it in as part of
  this chart; otherwise leave it `false` and the existing installation is
  reused automatically.

## Install

```sh
helm dependency update charts/rc-admission-webhook
helm install rc-admission-webhook charts/rc-admission-webhook \
  --namespace rc-admission-webhook --create-namespace
```

## Values

| Key | Description | Default |
| --- | --- | --- |
| `image.repository` / `image.tag` | Webhook image | `ghcr.io/runtimeconditions/rc-admission-webhook` / `latest` |
| `certManager.installDependency` | Install cert-manager as part of this release | `false` |
| `certManager.selfSigned` | Use a self-signed Issuer scoped to this release | `true` |
| `networkPolicy.enabled` | Restrict webhook ingress to namespaces labeled `webhook: enabled` | `true` |
