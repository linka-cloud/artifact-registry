{{- $repoType := "apk" -}}

# APK Packages

Publish [Alpine](https://pkgs.alpinelinux.org/) packages.

## Requirements

To work with the APK registry, you need either the `lkar` client or an HTTP client like `curl` to upload and finally, a
package manager like `apk` to install packages.

The following examples use `apk`.


### Variable used in the examples

| Placeholder         | Description                       |
|---------------------|-----------------------------------|
| `image`             | The oci image used as backend.    |
| `branch`            | The branch to use.                |
| `repository`        | The repository to use.            |
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


To setup the APK registry on the local machine, run the following command:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

```shell
lkar {{ $repoType }} setup {{ $repo }} <branch> <repository>
```

{{- end }}
{{- end }}

### curl

If the registry is private, provide credentials in the url:

```
https://<username>:<password_or_token>@<url>
```

To register the Alpine registry using the repository script, run the following command:

{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}

```shell
curl -s https://{{ $url }}/<branch>/<repository>/setup | sh
```

{{- end }}
{{- end }}

### Manually

If the registry is private, provide credentials in the url:

```
https://<username>:<password_or_token>@<url>
```

To register the Alpine registry add the url to the list of known apk sources (`/etc/apk/repositories`):

{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}

```
https://{{ $url }}/<branch>/<repository>
```

{{- end }}
{{- end }}


The Alpine registry files are signed with a RSA key which must be known to apk. 
From the `/etc/apk/keys/` directory, download the public key:

{{- range $deployMode := $.DeployModes }}

{{ if not $.DeployMode }}
#### {{ $deployMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode 0 $repoType "key" }}

```shell
curl -JO https://{{ $url }}/key
```

{{- end }}

Afterward, update the local package index:

```shell
apk update
```

## Publish a package


### lkar

To publish an APK package, run the following command:


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

```shell
lkar {{ $repoType }} push {{ $repo }} <branch> <repository> path/to/file.apk
```

{{- end }}
{{- end }}

### curl
    
To publish an APK package, perform an HTTP `PUT` operation with the package content in the request body.


{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $url := $.RegistryURL $deployMode $repoMode $repoType "<image>" }}
{{- $exampleURL := $.RegistryURL $deployMode $repoMode $repoType "user/image" }}

```
https://{{ $url }}/<branch>/<repository>/push
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.{{ $repoType }} \
     https://{{ $exampleURL }}/v3.17/main/push
```

{{- end }}
{{- end }}

## Delete a package

### lkar

To delete an APK package, run the following commands:

{{- range $deployMode := $.DeployModes }}
{{- range $repoMode := $.RepoModes }}

{{ if not $.RepoMode }}
#### {{ $deployMode }} {{ $repoMode }}
{{- end }}

{{- $repo := $.Registry $deployMode $repoMode $repoType "<image>" }}

First retrieve the path to package you want to delete:

```shell
lkar {{ $repoType }} ls {{ $repo }} <branch> <repository>
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

To delete an APK package, first retrieve the path to the package in the repository:

```
GET https://{{ $packagesURL }}
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://{{ $url }}/<branch>/<repository>/<architecture>/<filename>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://{{ $exampleURL }}/v3.17/main/test-package-1.0.0.apk
```

{{- end }}
{{- end }}


## Install a package

To install a package from the APK registry, execute the following commands:

```shell
# use latest version
apk add {package_name}
# use specific version
apk add {package_name}={package_version}
```
