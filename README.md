# rc-admission-webhook
An admissions webhook to validate Runtime Conditions Profiles when stored as a custom resource in Kubernetes

## Install

The webhook is installed with the Helm chart in [`charts/rc-admission-webhook`](charts/rc-admission-webhook),
which also lists the prerequisites (the `RuntimeConditionsProfile` CRD and cert-manager) and every
configurable value.

## Releases

Releases are cut by labeling a pull request before it is merged into `main`:

| Label | Effect |
| --- | --- |
| `release:major` / `release:minor` / `release:patch` | Bumps the latest `v*` tag (the first release is `v0.1.0`), creates a GitHub release, and pushes the image to `ghcr.io/runtimeconditions/rc-admission-webhook` as both `:<tag>` and `:latest` |
| none | Nothing is tagged or published |

The chart's default `image.tag` is `latest`, so installing it with default values only works once a
release has been cut and the package is public.
