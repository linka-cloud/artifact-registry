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



#### Subpath Single

```shell
lkar login artifact-registry.example.org
```



#### Subpath Multi

```shell
lkar login artifact-registry.example.org/<image>
```



#### Subdomain Single

```shell
lkar login rpm.example.org
```



#### Subdomain Multi

```shell
lkar login rpm.example.org/<image>
```

To setup the RPM registry on the local machine, run the following command:



#### Subpath Single

```shell
lkar rpm setup artifact-registry.example.org
```



#### Subpath Multi

```shell
lkar rpm setup artifact-registry.example.org/<image>
```



#### Subdomain Single

```shell
lkar rpm setup rpm.example.org
```



#### Subdomain Multi

```shell
lkar rpm setup rpm.example.org/<image>
```

### curl

If the registry is private, provide credentials in the url:

```
https://<username>:<password_or_token>@<url>
```

To register the repository using the generated script, run the following command:



#### Subpath Single

```shell
curl -s https://artifact-registry.example.org/rpm/setup | sh
```



#### Subpath Multi

```shell
curl -s https://artifact-registry.example.org/rpm/<image>/setup | sh
```



#### Subdomain Single

```shell
curl -s https://rpm.example.org/setup | sh
```



#### Subdomain Multi

```shell
curl -s https://rpm.example.org/<image>/setup | sh
```

### Manually

If the registry is private, provide credentials in the url:

```
https://{username}:{password_or_token}@<url>
```

### With `config-manager`



##### Subpath Single

*Not supported*



##### Subpath Multi

```shell
dnf config-manager --add-repo https://artifact-registry.example.org/rpm/<image>.repo
```



##### Subdomain Single

*Not supported*



##### Subdomain Multi

```shell
dnf config-manager --add-repo https://rpm.example.org/<image>.repo
```

### With repository file

To register the repository add the repository definition in the `/etc/yum.repos.d/` directory:



##### Subpath Single

```shell
curl https://artifact-registry.example.org/rpm/.repo | sudo tee /etc/yum.repos.d/example.repo
```



##### Subpath Multi

```shell
curl https://artifact-registry.example.org/rpm/<image>.repo | sudo tee /etc/yum.repos.d/example.repo
```



##### Subdomain Single

```shell
curl https://rpm.example.org/.repo | sudo tee /etc/yum.repos.d/example.repo
```



##### Subdomain Multi

```shell
curl https://rpm.example.org/<image>.repo | sudo tee /etc/yum.repos.d/example.repo
```

## Publish a package

### lkar

To publish an RPM package, run the following command:


#### Subpath Single

```shell
lkar rpm push artifact-registry.example.org path/to/file.rpm
```


#### Subpath Multi

```shell
lkar rpm push artifact-registry.example.org/<image> path/to/file.rpm
```


#### Subdomain Single

```shell
lkar rpm push rpm.example.org path/to/file.rpm
```


#### Subdomain Multi

```shell
lkar rpm push rpm.example.org/<image> path/to/file.rpm
```

### curl

To publish an RPM package, perform an HTTP `PUT` operation with the package content in the request body.


#### Subpath Single

```
https://artifact-registry.example.org/rpm/push
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.rpm \
     https://artifact-registry.example.org/rpm/push
```


#### Subpath Multi

```
https://artifact-registry.example.org/rpm/<image>/push
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.rpm \
     https://artifact-registry.example.org/rpm/user/image/push
```


#### Subdomain Single

```
https://rpm.example.org/push
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.rpm \
     https://rpm.example.org/push
```


#### Subdomain Multi

```
https://rpm.example.org/<image>/push
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token \
     --upload-file path/to/file.rpm \
     https://rpm.example.org/user/image/push
```

## Delete a package

### lkar

To delete an RPM package, run the following commands:


#### Subpath Single

First retrieve the path to package you want to delete:

```shell
lkar rpm ls artifact-registry.example.org
```

Then use the path to delete the package:

```shell
lkar rpm rm artifact-registry.example.org <path>
```


#### Subpath Multi

First retrieve the path to package you want to delete:

```shell
lkar rpm ls artifact-registry.example.org/<image>
```

Then use the path to delete the package:

```shell
lkar rpm rm artifact-registry.example.org/<image> <path>
```


#### Subdomain Single

First retrieve the path to package you want to delete:

```shell
lkar rpm ls rpm.example.org
```

Then use the path to delete the package:

```shell
lkar rpm rm rpm.example.org <path>
```


#### Subdomain Multi

First retrieve the path to package you want to delete:

```shell
lkar rpm ls rpm.example.org/<image>
```

Then use the path to delete the package:

```shell
lkar rpm rm rpm.example.org/<image> <path>
```

### curl


#### Subpath Single

To delete an RPM package, first retrieve the path to the package in the repository:

```
GET https://artifact-registry.example.org/_packages/rpm
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://artifact-registry.example.org/rpm/<filepath>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://artifact-registry.example.org/rpm/test-package-1.0.0.rpm
```


#### Subpath Multi

To delete an RPM package, first retrieve the path to the package in the repository:

```
GET https://artifact-registry.example.org/_packages/rpm/<image>
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://artifact-registry.example.org/rpm/<image>/<filepath>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://artifact-registry.example.org/rpm/user/image/test-package-1.0.0.rpm
```


#### Subdomain Single

To delete an RPM package, first retrieve the path to the package in the repository:

```
GET https://rpm.example.org/_packages
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://rpm.example.org/<filepath>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://rpm.example.org/test-package-1.0.0.rpm
```


#### Subdomain Multi

To delete an RPM package, first retrieve the path to the package in the repository:

```
GET https://rpm.example.org/_packages/<image>
```

Then perform an HTTP `DELETE` operation. This will delete the package version too if there is no
file left.

```
DELETE https://rpm.example.org/<image>/<filepath>
```

Example request using HTTP Basic authentication:

```shell
curl --user username:password_or_token -X DELETE \
     https://rpm.example.org/user/image/test-package-1.0.0.rpm
```


## Install a package

To install a package from the RPM registry, execute the following commands:

```shell
# use latest version
yum install {package_name}
# use specific version
yum install {package_name}={package_version}
```
