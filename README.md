<p align="center">
    <img alt="LK Artifact Registry" src="ui/src/img/lkar-no-background.png" width='720px'/>
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

- deb
- rpm
- apk
- helm
- ... (more to come)

## Quick start (evaluation only)

Let's start by deploying a docker registry (without authentication), the artifact registry and an alpine based client to test it:

```yaml
version: '3.7'
services:
  registry:
    image: registry:2
  artifact-registry:
    container_name: artifact-registry
    image: linkacloud/artifact-registry:dev
    environment:
      ARTIFACT_REGISTRY_AES_KEY: "noop"
    command:
    - --backend=registry:5000
    - --no-https
    ports:
    - "9887:9887"
  artifact-registry-client:
    container_name: artifact-registry-client
    image: linkacloud/artifact-registry:dev
    init: true
    entrypoint: ["/bin/sh", "-c"]
    command:
    - sleep infinity
```

```bash
curl https://raw.githubusercontent.com/linka-cloud/artifact-registry/main/docker-compose.yaml | docker compose -f - up -d
```

Now, let's get some packages:

```bash
docker exec -it artifact-registry-client sh -c "apk fetch --no-cache -o /tmp -R curl jq"
```

Upload them to the registry:

```bash
docker exec -it artifact-registry-client sh -c "find /tmp -name '*.apk' -exec lkar apk push --plain-http artifact-registry:9887/test v3.18 main {} \;"
# remove any existing apk repository and cache
docker exec -it artifact-registry-client sh -c "rm -rf /tmp/* && rm -rf /var/cache/apk/* && rm -rf /etc/apk/repositories"
```

Setup artifact-registry as repository:

```bash
docker exec -it artifact-registry-client sh -c "lkar apk setup --plain-http artifact-registry:9887/test v3.18 main"
```

And finally, install the packages:

```bash
docker exec -it artifact-registry-client sh -c "apk add --no-cache curl jq"
```


Clean up:

```bash
curl https://raw.githubusercontent.com/linka-cloud/artifact-registry/main/docker-compose.yaml | docker compose -f - down --volumes --remove-orphans
```

## Getting started

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


<!--- ### Using the registry --->

<!--- TODO(adphi): add instructions for installing the client --->


<!--- TODO(adphi): add lkard and lkar usage --->


## Acknowledgements

This package formats implementations are based on the amazing work done of the [Gitea](https://gitea.io) team.

Many thanks to them for their work, especially to [@KN4CK3R](https://github.com/KN4CK3R).
