{{- $repoType := "rpm" -}}

# RPM Packages

Publish rpm packages.

## Requirements

To work with the RPM registry, you need either the `lkar` client or an HTTP client like `curl` to upload and finally, a
package manager like `yum` or `dnf` to install packages.

The following examples use mostly `yum`.

### Variable used in the examples

| Placeholder         | Description                                       |
|---------------------|---------------------------------------------------|
| `image`             | The oci image used as backend.                    |
| `username`          | The repository user.                              |
| `password_or_token` | The repository password or token.                 |
| `architecture`      | The package architecture.                         |
| `filepath`          | The path in the repository to the file to delete. |

## Configuring the package registry

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

To setup the RPM registry on the local machine, run the following command:

{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}

#### {{ $deployMode }} {{ $repoMode }}

{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

```shell
lkar {{ $repoType }} setup {{ $repo }}
```

{{- end }}
{{- end }}

### curl

If the registry is private, provide credentials in the url:

```
https://<username>:<password_or_token>@<url>
```

To register the repository using the generated script, run the following command:

{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}

#### {{ $deployMode }} {{ $repoMode }}

{{- end }}

{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}

```shell
curl -s https://{{ $url }}/setup | sh
```

{{- end }}
{{- end }}

### Manually

If the registry is private, provide credentials in the url:

```
https://{username}:{password_or_token}@<url>
```

### With `config-manager`

{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}

##### {{ $deployMode }} {{ $repoMode }}

{{- end }}

{{- if eq $repoMode.String "Multi" }}

{{- $url := $.APIEndpoint $deployMode $repoMode $repoType "<image>" "" }}

```shell
dnf config-manager --add-repo https://{{ $url }}.repo
```

{{- else }}

*Not supported*

{{- end }}

{{- end }}
{{- end }}

### With repository file

To register the repository add the repository definition in the `/etc/yum.repos.d/` directory:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}

##### {{ $deployMode }} {{ $repoMode }}

{{- end }}

{{- $url := $.APIEndpoint $deployMode $repoMode $repoType "<image>" "" }}

```shell
curl https://{{ $url }}{{ if eq $repoMode.String "Single" }}/{{ end }}.repo | sudo tee /etc/yum.repos.d/example.repo
```

{{- end }}
{{- end }}

## Publish a package

### lkar

To publish an RPM package, run the following command:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

```shell
lkar {{ $repoType }} push {{ $repo }} path/to/file.rpm
```

{{- end }}
{{- end }}

### curl

To publish an RPM package, perform an HTTP `PUT` operation with the package content in the request body.


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}
{{- $exampleURL := $.RegistryURL $deployMode $repoMode $repoType "user/image" }}

```
https://{{ $url }}/push
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.{{ $repoType }} \
     https://{{ $exampleURL }}/push
```

{{- end }}
{{- end }}

## Delete a package

### lkar

To delete an RPM package, run the following commands:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

First retrieve the path to package you want to delete:

```shell
lkar {{ $repoType }} ls {{ $repo }}
```

Then use the path to delete the package:

```shell
lkar {{ $repoType }} rm {{ $repo }} <path>
```

{{- end }}
{{- end }}

### curl


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $packagesURL := $.APIEndpoint $deployMode $repoMode $repoType "<image>" "_packages"}}
{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}
{{- $exampleURL := $.RegistryURL $deployMode $repoMode $repoType "user/image" }}

To delete an RPM package, first retrieve the path to the package in the repository:

```
GET https://{{ $packagesURL }}
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://{{ $url }}/<filepath>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://{{ $exampleURL }}/test-package-1.0.0.rpm
```

{{- end }}
{{- end }}


## Install a package

To install a package from the RPM registry, execute the following commands:

```shell
# use latest version
yum install {package_name}
# use specific version
yum install {package_name}={package_version}
```
