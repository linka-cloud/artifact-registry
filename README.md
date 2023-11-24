<p align="center">
    <img alt="LK Artifact Registry" src="docs/assets/lkar-no-background.png" width='720px'/>
</p>


# LK Artifact Registry

[![PkgGoDev](https://pkg.go.dev/badge/go.linka.cloud/artifact-registry)](https://pkg.go.dev/go.linka.cloud/artifact-registry)
[![Go Report Card](https://goreportcard.com/badge/go.linka.cloud/artifact-registry)](https://goreportcard.com/report/go.linka.cloud/artifact-registry)

*Distribute your artifacts to your end users without any additional administration or maintenance costs.*

Artifact Registry is a 100% stateless enterprise ready artifact registry.

It uses any compatible oci-registry as backend, for both storage, authentication and authorization, making it easy to deploy and maintain.

It can host as many repositories as you want, all being backed by a single oci-repository (docker image).

It has two main parts:
- lkard: the registry server which expose a small web-ui
- lkar: the command line client

## Packages formats

The following package formats are supported:

- <img alt="deb packages" style="vertical-align:middle" src="docs/assets/deb.svg" width='32px'/> <a href='docs/packages/deb.md'>deb</a>

- <img alt="rpm packages" style="vertical-align:middle" src="docs/assets/rpm.png" width='32px'/> <a href='docs/packages/rpm.md'>rpm</a>

- <img alt="apk packages" style="vertical-align:middle" src="docs/assets/apk.png" width='32px'/> <a href='docs/packages/apk.md'>apk</a>

- <img alt="helm packages" style="vertical-align:middle" src="docs/assets/helm.svg" width='32px'/> <a href='docs/packages/helm.md'>helm</a>

- ... more to come

## Features

### Deployment Modes

The registry can be configured in different modes:

- Multi Repository Mode (default):

  The multi-repositories mode uses one oci-image per repository.
  It is useful when you are you want to have a different repository for each of your projects.


- Single Repository Mode:

  The single-repository mode uses only one oci-image as storage backend.
  It is useful when you want to distribute all your packages from a single place.

  > To configure this mode, you need to set the `lkard --repo` flag or the `config.backend.repo` helm value to the name of the repository you want to use.

It can also be configured to serve the repositories as sub-path or sub-domain.

- Sub-path Mode (default):
  The sub-path mode uses a different sub-path for each repository types.
  For example, the *deb* repository will be served from `example.com/deb` and the *rpm* repository from `example.com/rpm`.


- Sub-domain Mode:

  The sub-domain mode uses a different sub-domain for each repository types.
  For example, the *deb* repository will be served from `deb.example.com` and the *rpm* repository from `rpm.example.com`.

  > To configure this mode, you need to set the `lkard --domain` flag or the `config.domain` helm value to the domain name you want to use
  and create the DNS entries pointing to the registry.

### Registry Proxy support

The artifact-registry has built-in support for registry proxies.

> ⚠️ If you intend to use the registry with `docker.io` as backend, it is highly recommended to use a registry pull-through cache/proxy like [docker.io/registry](https://hub.docker.com/_/registry) or [harbor](https://goharbor.io/)...
> otherwise you can be sure that the artifact-registry ip will be banned.


Command line:

    The proxy is configuratble using the following flags:

    ```
    --proxy string             proxy backend registry hostname (and port if not 443 or 80) [$ARTIFACT_REGISTRY_PROXY]
    --proxy-client-ca string   proxy tls client certificate authority [$ARTIFACT_REGISTRY_PROXY_CLIENT_CA]
    --proxy-insecure           disable proxy registry client tls verification [$ARTIFACT_REGISTRY_PROXY_INSECURE]
    --proxy-no-https           disable proxy registry client https [$ARTIFACT_REGISTRY_PROXY_NO_HTTPS]
    --proxy-password string    proxy registry password [$ARTIFACT_REGISTRY_PROXY_PASSWORD]
    --proxy-user string        proxy registry user [$ARTIFACT_REGISTRY_PROXY_USER]
    ```

Helm:

    The proxy is configuratble using the following helm values:

    | Key                                        | Description                 |
    |--------------------------------------------|-----------------------------|
    | config.proxy.host                          | Proxy hostname              |
    | config.proxy.insecure                      | Disable proxy TLS verify    |
    | config.proxy.plainHTTP                     | Use HTTP for proxy          |
    | config.proxy.clientCA                      | Proxy CA secret             |
    | config.proxy.username                      | Proxy username              |
    | config.proxy.password                      | Proxy password              |


For more information, see the [lkard reference](docs/reference/lkard/lkard.md) and the [helm chart's README](helm/artifact-registry/README.md).


## Getting started

### Evaluating the registry

See the [Getting Started](./docs/getting-started.md) guide for a quick introduction to the registry.

### Deploying the registry

Deploy the registry using helm:

```bash
helm repo add linka-cloud https://helm.linka.cloud

REGISTRY=registry.example.org

helm upgrade \
    --install \
    --create-namespace \
    --namespace artifact-registry \
    --set config.backend.host=$REGISTRY \
    artifact-registry \
    linka-cloud/artifact-registry
```

See the [Chart's README](./helm/artifact-registry/README.md) for the available configuration options.


<!--- TODO(adphi): add instructions for installing the client --->


### Using the registry

See the [documentation](docs/README.md) for more information about the registry usage.

<!--- TODO(adphi): no seriously... add some minimal presentation... --->


## Acknowledgements

This package formats implementations are based on the amazing work done of the [Gitea](https://gitea.io) team.

Many thanks to them for their work, especially to [@KN4CK3R](https://github.com/KN4CK3R).
