# Helm Chart Registry

Publish [Helm](https://helm.sh/) charts for your users or organization.

## Requirements

To work with the Helm Chart registry use a `lkar` or simple HTTP client like `curl`.

### Variable used in the examples

| Placeholder         | Description                       |
|---------------------|-----------------------------------|
| `image`             | The oci image used as backend.    |
| `username`          | The repository user.              |
| `password_or_token` | The repository password or token. |


## Publish a package

Publish a package by running the following command:

### lkar

If the registry is private, start by log in the registry:

#### Subpath Single

```
lkar login artifact-registry.example.org
```

#### Subpath Multi

```
lkar login artifact-registry.example.org/<image>
```

#### Subdomain Single

```
lkar login helm.example.org
```

#### Subdomain Multi

```
lkar login helm.example.org/<image>
```

You can then publish a chart by running the following command:

```
lkar helm push artifact-registry.example.org path/to/file.tgz
```

#### Subpath Multi

```
lkar helm push artifact-registry.example.org/<image> path/to/file.tgz
```

#### Subdomain Single

```
lkar helm push helm.example.org path/to/file.tgz
```

#### Subdomain Multi

```
lkar helm push helm.example.org/<image> path/to/file.tgz
```

### curl

To publish a helm Chart, perform an HTTP `PUT` operation with the package content in the request body.

#### Subpath Single

```shell
curl --user <username>:<password_or_token> -X PUT --upload-file path/to/file.tgz https://artifact-registry.example.org/helm/push
```

#### Subpath Multi

```shell
curl --user <username>:<password_or_token> -X PUT --upload-file path/to/file.tgz https://artifact-registry.example.org/helm/<image>/push
```

#### Subdomain Single

```shell
curl --user <username>:<password_or_token> -X PUT --upload-file path/to/file.tgz https://helm.example.org/push
```

#### Subdomain Multi

```shell
curl --user <username>:<password_or_token> -X PUT --upload-file path/to/file.tgz https://helm.example.org/<image>/push
```

## Install a package

To install a Helm char from the registry, start by adding the repository to your Helm client:

#### Subpath Single

```shell
helm repo add  --username <username> --password <password> example https://artifact-registry.example.org/helm
```

#### Subpath Multi

```shell
helm repo add  --username <username> --password <password> example https://artifact-registry.example.org/<image>/helm
```

#### Subdomain Single

```shell
helm repo add  --username <username> --password <password> example https://helm.example.org
```

#### Subdomain Multi

```shell
helm repo add  --username <username> --password <password> example https://helm.example.org/<image>
```

You can then install a chart by running the following command:

```shell
helm repo update
helm install my-chart example/my-chart
```
