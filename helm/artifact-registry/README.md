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

| Key                                        | Description                                                           | Default                      |
|--------------------------------------------|-----------------------------------------------------------------------|------------------------------|  
| replicaCount                               | Number of replicas                                                    | 1                            |
| image.repository                           | Image repository                                                      | linkacloud/artifact-registry |
| image.pullPolicy                           | Image pull policy                                                     | IfNotPresent                 |
| image.tag                                  | Image tag override                                                    | ""                           |
| imagePullSecrets                           | Image pull secrets                                                    | []                           |  
| nameOverride                               | Override name                                                         | ""                           |
| fullnameOverride                           | Override fullname                                                     | ""                           |
| config.aes.key                             | AES encryption key                                                    | ""                           |
| config.aes.secret                          | AES encryption secret name                                            | ""                           |
| config.aes.secret.name                     | Existing AES encryption secret name                                   | ""                           |
| config.aes.secret.key                      | Existing AES encryption secret key                                    | ""                           |
| config.disableUI                           | Disable web UI                                                        | false                        |
| config.backend.host                        | Backend registry hostname                                             | docker.io                    |  
| config.backend.repo                        | Backend registry repository                                           | ""                           |
| config.backend.insecure                    | Disable backend TLS verify                                            | false                        | 
| config.backend.plainHTTP                   | Use HTTP for backend                                                  | false                        |
| config.backend.clientCA                    | Backend CA secret                                                     | ""                           |
| config.proxy.host                          | Proxy hostname                                                        | ""                           |
| config.proxy.insecure                      | Disable proxy TLS verify                                              | false                        |
| config.proxy.plainHTTP                     | Use HTTP for proxy                                                    | false                        |
| config.proxy.clientCA                      | Proxy CA secret                                                       | ""                           |
| config.proxy.username                      | Proxy username                                                        | ""                           |
| config.proxy.password                      | Proxy password                                                        | ""                           |
| config.domain                              | Domain name                                                           | ""                           |
| config.port                                | Listening port                                                        | 9887                         |
| config.tagArtifacts                        | Tag artifacts                                                         | false                        |
| config.tls.secretName                      | TLS secret name                                                       | ""                           |
| env                                        | Environment variables                                                 | []                           |
| serviceAccount.create                      | Create service account                                                | false                        |  
| serviceAccount.annotations                 | Service account annotations                                           | {}                           |
| serviceAccount.name                        | Service account name                                                  | ""                           |
| podAnnotations                             | Pod annotations                                                       | {}                           |
| podSecurityContext                         | Pod security context                                                  | {}                           |
| securityContext                            | Container security context                                            | {}                           |
| service.type                               | Service type                                                          | ClusterIP                    |
| service.port                               | Service port                                                          | 9887                         | 
| service.annotations                        | Service annotations                                                   | {}                           |
| ingress.enabled                            | Enable ingress                                                        | false                        |
| ingress.className                          | Ingress class name                                                    | ""                           |
| ingress.annotations                        | Ingress annotations                                                   | {}                           |
| ingress.hosts                              | Ingress hosts                                                         | []                           |
| ingress.tls                                | Ingress TLS config                                                    | []                           |
| resources                                  | Resource limits and requests                                          | {}                           |
| autoscaling.enabled                        | Enable autoscaling                                                    | false                        |  
| autoscaling.minReplicas                    | Autoscaling min replicas                                              | 1                            |
| autoscaling.maxReplicas                    | Autoscaling max replicas                                              | 100                          |
| autoscaling.targetCPUUtilizationPercentage | Target CPU usage for autoscaling                                      | 80                           | 
| nodeSelector                               | Node selector                                                         | {}                           |
| tolerations                                | Tolerations                                                           | []                           |
| affinity                                   | Affinity                                                              | {}                           |
| registry.enabled                           | Deploy the docker registry Chart                                      | true                         |
| extraTemplates                             | extraTemplates is an Array of extra manifests or templates to deploy. | []                           |
