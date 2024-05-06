{{- $repoType := "helm" -}}

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

{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

```shell
lkar login {{ $repo }}
```

{{- end }}
{{- end }}

You can then publish a chart by running the following command:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

```shell
lkar {{ $repoType }} push {{ $repo }} path/to/file.tgz
```

{{- end }}
{{- end }}

### curl

To publish a helm Chart, perform an HTTP `PUT` operation with the package content in the request body.

{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}

```shell
curl --user <username>:<password_or_token> -X PUT --upload-file path/to/file.tgz https://{{ $url }}/push
```

{{- end }}
{{- end }}

## Install a package

To install a Helm char from the registry, start by adding the repository to your Helm client:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}

```shell
helm repo add  --username <username> --password <password> example https://{{ $url }}
```

{{- end }}
{{- end }}

You can then install a chart by running the following command:

```shell
helm repo update
helm install my-chart example/my-chart
```
