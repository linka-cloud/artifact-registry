{{- $repoType := "deb" -}}

# DEB Packages

Publish deb packages.

## Requirements

To work with the DEB registry, you need either the `lkar` client or an HTTP client like `curl` to upload and finally, a
package manager like `apt` to install packages.

The following examples use `apt`.


### Variable used in the examples

| Placeholder         | Description                       |
|---------------------|-----------------------------------|
| `image`             | The oci image used as backend.    |
| `distribution`      | The distriution to use.           |
| `component`         | The component to use.             |
| `username`          | The repository user.              |
| `password_or_token` | The repository password or token. |
| `architecture`      | The package architecture.         |
| `filepath`          | The path to the file to delete.   |

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


To setup the DEB registry on the local machine, run the following command:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

```shell
lkar {{ $repoType }} setup {{ $repo }} <distribution> <component>
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
curl -s https://{{ $url }}/<distribution>/<component>/setup | sh
```

{{- end }}
{{- end }}

### Manually

If the registry is private, provide credentials in the url:

```
https://{username}:{password_or_token}@<url>
```

To register the repository add the url to the list of known deb sources (`/etc/apt/sources.list`):


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}

```
echo "deb https://{{ $url }} <distribution> <component>" | sudo tee -a /etc/apt/sources.list
```

{{- end }}
{{- end }}

The registry files are signed with a GPG key which must be known to apt.


{{- range $deployMode := $.DeployModes }}

{{ if not $.DeployMode }}
#### {{ $deployMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode 0 $repoType "key" }}

```shell
sudo curl  https://{{ $url }}/repository.key -o /etc/apt/trusted.gpg.d/example.asc
```

{{- end }}

Afterward, update the local package index:

```shell
apt update
```

## Publish a package


### lkar

To publish an DEB package, run the following command:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

```shell
lkar {{ $repoType }} push {{ $repo }} <distribution> <component> path/to/file.deb
```

{{- end }}
{{- end }}

### curl

To publish an DEB package, perform an HTTP `PUT` operation with the package content in the request body.


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}
{{- $exampleURL := $.RegistryURL $deployMode $repoMode $repoType "user/image" }}

```
https://{{ $url }}/<distribution>/<component>/push
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.{{ $repoType }} \
     https://{{ $exampleURL }}/focal/main/push
```

{{- end }}
{{- end }}


## Delete a package

### lkar

To delete an DEB package, run the following commands:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

First retrieve the path to package you want to delete:

```shell
lkar {{ $repoType }} ls {{ $repo }} <distribution> <component>
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

To delete an DEB package, first retrieve the path to the package in the repository:

```
GET https://{{ $packagesURL }}
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://{{ $url }}/pool/<distribution>/<component>/<architecture>/<filename>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://{{ $exampleURL }}/pool/focal/main/test-package-1.0.0.deb
```

{{- end }}
{{- end }}



## Install a package

To install a package from the DEB registry, execute the following commands:

```shell
# use latest version
apt install {package_name}
# use specific version
apt install {package_name}={package_version}
```
