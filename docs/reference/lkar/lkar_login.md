## lkar login

Login to an Artifact Registry repository

```
lkar login [registry] [flags]
```

### Examples

```

Log in with username and password from command line flags:
  lkar login -u username -p password localhost:5000

Log in with username and password from stdin:
  lkar login -u username --password-stdin localhost:5000

Log in with username and password in an interactive terminal and no TLS check:
  lkar login --insecure localhost:5000

```

### Options

```
  -h, --help             help for login
      --password-stdin   Take the password from stdin
```

### Options inherited from parent commands

```
      --ca-file string   CA certificate file
  -d, --debug            Enable debug logging
  -k, --insecure         Do not verify tls certificates
  -p, --pass string      Password
  -H, --plain-http       Use http instead of https
  -u, --user string      Username
```

### SEE ALSO

* [lkar](lkar.md)	 - An OCI based Artifact Registry

