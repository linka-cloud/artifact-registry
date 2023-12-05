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
lkar login deb.example.org
```

#### Subdomain Multi

```
lkar login deb.example.org/<image>
```


To setup the DEB registry on the local machine, run the following command:

#### Subpath Single

```
lkar deb setup artifact-registry.example.org <distribution> <component>
```

#### Subpath Multi

```
lkar deb setup artifact-registry.example.org/<image> <distribution> <component>
```

#### Subdomain Single

```
lkar deb setup deb.example.org <distribution> <component>
```

#### Subdomain Multi

```
lkar deb setup deb.example.org/<image> <distribution> <component>
```

### curl

If the registry is private, provide credentials in the url:

```
https://{username}:{password_or_token}@<url>
```

To register the repository using the generated script, run the following command:

#### Subpath Single

```
curl -s https://artifact-registry.example.org/deb/<distribution>/<component>/setup | sh
```

#### Subpath Multi

```
curl -s https://artifact-registry.example.org/<image>/deb/<distribution>/<component>/setup | sh
```

#### Subdomain Single

```
curl -s https://deb.example.org/<distribution>/<component>/setup | sh
```

#### Subdomain Multi

```
curl -s https://deb.example.org/<image>/<distribution>/<component>/setup | sh
```

### Manually

If the registry is private, provide credentials in the url:

```
https://{username}:{password_or_token}@<url>
```

To register the repository add the url to the list of known deb sources (`/etc/apt/sources.list`):

#### Subpath Single

```
echo "deb https://artifact-registry.example.org/deb <distribution> <component>" | sudo tee -a /etc/apt/sources.list
```

#### Subpath Multi

```
echo "deb https://artifact-registry.example.org/<image>/deb <distribution> <component>" | sudo tee -a /etc/apt/sources.list
```

#### Subdomain Single

```
echo "deb https://deb.example.org <distribution> <component>" | sudo tee -a /etc/apt/sources.list
```

#### Subdomain Multi

```
echo "deb https://deb.example.org/<image> <distribution> <component>" | sudo tee -a /etc/apt/sources.list
```


The registry files are signed with a GPG key which must be known to apt.

#### Subpath

```shell
sudo curl  https://artifact-registry.example.org/deb/repository.key -o /etc/apt/trusted.gpg.d/example.asc
```

#### Subdomain

```shell
sudo curl  https://deb.example.org/repository.key -o /etc/apt/trusted.gpg.d/example.asc
```

Afterward, update the local package index:

```shell
apt update
```

## Publish a package


### lkar

To publish an DEB package, run the following command:

#### Subpath Single

```
lkar deb push artifact-registry.example.org <distribution> <component> path/to/file.deb
```

#### Subpath Multi

```
lkar deb push artifact-registry.example.org/<image> <distribution> <component> path/to/file.deb
```

#### Subdomain Single

```
lkar deb push deb.example.org <distribution> <component> path/to/file.deb
```

#### Subdomain Multi

```
lkar deb push deb.example.org/<image> <distribution> <component> path/to/file.deb
```

### curl

To publish an DEB package, perform an HTTP `PUT` operation with the package content in the request body.


#### Subpath Single

```
https://artifact-registry.example.org/deb/<distribution>/<component>/push
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.deb \
     https://artifact-registry.example.org/deb/focal/main
```

#### Subpath Multi

```
https://artifact-registry.example.org/<image>/deb/<distribution>/<component>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.deb \
     https://artifact-registry.example.org/user/image/deb/focal/main
```


#### Subdomain Single

```
https://deb.example.org/<distribution>/<component>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.deb \
     https://deb.example.org/focal/main
```

#### Subdomain Multi

```
https://deb.example.org/<image>/<distribution>/<component>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.deb \
     https://deb.example.org/user/image/focal/main
```

## Delete a package

### lkar

To delete an DEB package, run the following command:

#### Subpath Single

First retrieve the path to package you want to delete:

```shell
lkar deb ls deb.example.org <distribution> <component>
```

Then use the path to delete the package:

```shell
lkar deb rm artifact-registry.example.org <path>
```

#### Subpath Multi

First retrieve the path to package you want to delete:

```shell
lkar deb ls artifact-registry.example.org/<image> <distribution> <component>
```

Then use the path to delete the package:

```shell
lkar deb rm artifact-registry.example.org/<image> <path>
```

#### Subdomain Single

First retrieve the path to package you want to delete:

```shell
lkar deb ls deb.example.org <distribution> <component>
```

Then use the path to delete the package:

```shell
lkar deb rm deb.example.org <path>
```

#### Subdomain Multi

First retrieve the path to package you want to delete:

```shell
lkar deb ls deb.example.org/<image> <distribution> <component>
```

Then use the path to delete the package:

```shell
lkar deb rm deb.example.org/<image> <path>
```

### curl


#### Subpath Single

To delete an DEB package, first retrieve the path to the package in the repository:

```
GET https://artifact-registry.example.org/_packages/deb
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://artifact-registry.example.org/deb/pool/{distribution}/{component}/{architecture}/{filename}
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://artifact-registry.example.org/deb/pool/focal/main/test-package-1.0.0.deb
```

#### Subpath Multi

To delete an DEB package, first retrieve the path to the package in the repository:

```
GET https://artifact-registry.example.org/_packages/deb/{image}
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://artifact-registry.example.org/{image}/deb/pool/{distribution}/{component}/{architecture}/{filename}
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://artifact-registry.example.org/deb/user/image/pool/focal/main/test-package-1.0.0.deb
```

#### Subdomain Single

To delete an DEB package, first retrieve the path to the package in the repository:

```
GET https://deb.example.org/_packages
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://deb.example.org/pool/{distribution}/{component}/{architecture}/{filename}
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://deb.example.org/pool/focal/main/test-package-1.0.0.deb
```

#### Subdomain Multi

To delete an DEB package, first retrieve the path to the package in the repository:

```
GET https://deb.example.org/_packages/{image}
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://deb.example.org/{image}/pool/{distribution}/{component}/{architecture}/{filename}
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://deb.example.org/user/image/pool/focal/main/test-package-1.0.0.deb
```

## Install a package

To install a package from the DEB registry, execute the following commands:

```shell
# use latest version
apt install {package_name}
# use specific version
apt install {package_name}={package_version}
```
