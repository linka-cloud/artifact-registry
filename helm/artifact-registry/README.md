# LK Artifact Registry Helm Chart

This Helm chart installs LK Artifact Registry on Kubernetes.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2+

## Installing the Chart

Fetch the latest Helm chart from the repository:

```bash
helm repo add linka-cloud https://helm.linka.cloud
```

To install the chart with the release name `artifact-registry`:

```bash
REGISTRY=registry.example.org
helm upgrade --install --set config.backend.host=$REGISTRY artifact-registry linka-cloud/artifact-registry
```

# Default values for lk-artifact-registry.

For more information please refer to the [values.yaml](values.yaml) file.
