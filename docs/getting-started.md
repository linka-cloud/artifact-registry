# Getting Started

## Deploy sample with docker-compose

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

## Push packages to the artifact-registry

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

## Setup artifact-registry as repository

```bash
docker exec -it artifact-registry-client sh -c "lkar apk setup --plain-http artifact-registry:9887/test v3.18 main"
```

## Install packages

And finally, install the packages:

```bash
docker exec -it artifact-registry-client sh -c "apk add --no-cache curl jq"
```


## Clean up:

```bash
curl https://raw.githubusercontent.com/linka-cloud/artifact-registry/main/docker-compose.yaml | docker compose -f - down --volumes --remove-orphans
```
