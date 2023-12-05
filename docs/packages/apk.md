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
lkar login apk.example.org
```

#### Subdomain Multi

```
lkar login apk.example.org/<image>
```


To setup the APK registry on the local machine, run the following command:

#### Subpath Single

```
lkar apk setup artifact-registry.example.org <branch> <repository>
```

#### Subpath Multi

```
lkar apk setup artifact-registry.example.org/<image> <branch> <repository>
```

#### Subdomain Single

```
lkar apk setup apk.example.org <branch> <repository>
```

#### Subdomain Multi

```
lkar apk setup apk.example.org/<image> <branch> <repository>
```

### curl

If the registry is private, provide credentials in the url:

```
https://{username}:{password_or_token}@<url>
```

To register the Alpine registry using the repository script, run the following command:

#### Subpath Single

```
curl -s https://artifact-registry.example.org/apk/<branch>/<repository>/setup | sh
```

#### Subpath Multi

```
curl -s https://artifact-registry.example.org/<image>/apk/<branch>/<repository>/setup | sh
```

#### Subdomain Single

```
curl -s https://apk.example.org/<branch>/<repository>/setup | sh
```

#### Subdomain Multi

```
curl -s https://apk.example.org/<image>/<branch>/<repository>/setup | sh
```

### Manually

If the registry is private, provide credentials in the url:

```
https://{username}:{password_or_token}@<url>
```

To register the Alpine registry add the url to the list of known apk sources (`/etc/apk/repositories`):

#### Subpath Single

```
https://artifact-registry.example.org/apk/<branch>/<repository>
```

#### Subpath Multi

```
https://artifact-registry.example.org/<image>/apk/<branch>/<repository>
```

#### Subdomain Single

```
https://apk.example.org/<branch>/<repository>
```

#### Subdomain Multi

```
https://apk.example.org/<image>/<branch>/<repository>
```


The Alpine registry files are signed with a RSA key which must be known to apk. 
From the `/etc/apk/keys/` directory, download the public key:

#### Subpath

```shell
curl -JO https://artifact-registry.example.org/apk/key
```

#### Subdomain

```shell
curl -JO https://apk.example.org/key
```

Afterward, update the local package index:

```shell
apk update
```

## Publish a package


### lkar

To publish an APK package, run the following command:

#### Subpath Single

```
lkar apk push artifact-registry.example.org <branch> <repository> path/to/file.apk
```

#### Subpath Multi

```
lkar apk push artifact-registry.example.org/<image> <branch> <repository> path/to/file.apk
```

#### Subdomain Single

```
lkar apk push apk.example.org <branch> <repository> path/to/file.apk
```

#### Subdomain Multi

```
lkar apk push apk.example.org/<image> <branch> <repository> path/to/file.apk
```

### curl
    
To publish an APK package, perform an HTTP `PUT` operation with the package content in the request body.


#### Subpath Single

```
https://artifact-registry.example.org/apk/<branch>/<repository>/push
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.apk \
     https://artifact-registry.example.org/apk/v3.17/main
```

#### Subpath Multi

```
https://artifact-registry.example.org/<image>/apk/<branch>/<repository>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.apk \
     https://artifact-registry.example.org/user/image/apk/v3.17/main
```


#### Subdomain Single

```
https://apk.example.org/<branch>/<repository>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.apk \
     https://apk.example.org/v3.17/main
```

#### Subdomain Multi

```
https://apk.example.org/<image>/<branch>/<repository>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.apk \
     https://apk.example.org/user/image/v3.17/main
```

## Delete a package

### lkar

To delete an APK package, run the following command:

#### Subpath Single

First retrieve the path to package you want to delete:

```shell
lkar apk ls apk.example.org <branch> <repository>
```

Then use the path to delete the package:

```shell
lkar apk rm artifact-registry.example.org <path>
```

#### Subpath Multi

First retrieve the path to package you want to delete:

```shell
lkar apk ls artifact-registry.example.org/<image> <branch> <repository>
```

Then use the path to delete the package:

```shell
lkar apk rm artifact-registry.example.org/<image> <path>
```

#### Subdomain Single

First retrieve the path to package you want to delete:

```shell
lkar apk ls apk.example.org <branch> <repository>
```

Then use the path to delete the package:

```shell
lkar apk rm apk.example.org <path>
```

#### Subdomain Multi

First retrieve the path to package you want to delete:

```shell
lkar apk ls apk.example.org/<image> <branch> <repository>
```

Then use the path to delete the package:

```shell
lkar apk rm apk.example.org/<image> <path>
```

### curl


#### Subpath Single

To delete an APK package, first retrieve the path to the package in the repository:

```
GET https://artifact-registry.example.org/_packages/apk
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://artifact-registry.example.org/apk/{branch}/{repository}/{architecture}/{filename}
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://artifact-registry.example.org/apk/v3.17/main/test-package-1.0.0.apk
```

#### Subpath Multi

To delete an APK package, first retrieve the path to the package in the repository:

```
GET https://artifact-registry.example.org/_packages/apk/{image}
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://artifact-registry.example.org/{image}/apk/{branch}/{repository}/{architecture}/{filename}
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://artifact-registry.example.org/apk/user/image/v3.17/main/test-package-1.0.0.apk
```

#### Subdomain Single

To delete an APK package, first retrieve the path to the package in the repository:

```
GET https://apk.example.org/_packages
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://apk.example.org/{branch}/{repository}/{architecture}/{filename}
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://apk.example.org/v3.17/main/test-package-1.0.0.apk
```

#### Subdomain Multi

To delete an APK package, first retrieve the path to the package in the repository:

```
GET https://apk.example.org/_packages/{image}
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://apk.example.org/{image}/{branch}/{repository}/{architecture}/{filename}
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://apk.example.org/user/image/v3.17/main/test-package-1.0.0.apk
```

## Install a package

To install a package from the APK registry, execute the following commands:

```shell
# use latest version
apk add {package_name}
# use specific version
apk add {package_name}={package_version}
```
