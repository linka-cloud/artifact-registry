## lkard

An OCI based Artifact Registry

```
lkard (repository) [flags]
```

### Options

```
      --addr string              address to listen on [$ARTIFACT_REGISTRY_ADDRESS] (default ":9887")
      --aes-key string           AES key to encrypt the repositories keys [$ARTIFACT_REGISTRY_AES_KEY]
      --backend string           registry backend hostname (and port if not 443 or 80) [$ARTIFACT_REGISTRY_BACKEND] (default "docker.io")
      --client-ca string         tls client certificate authority [$ARTIFACT_REGISTRY_CLIENT_CA]
  -d, --debug                    enable debug logging
      --disable-ui               disable the Web UI [$ARTIFACT_REGISTRY_DISABLE_UI]
      --domain string            domain to use to serve the repositories as subdomains [$ARTIFACT_REGISTRY_DOMAIN]
  -h, --help                     help for lkard
      --insecure                 disable backend registry client tls verification [$ARTIFACT_REGISTRY_INSECURE]
      --no-https                 disable backend registry client https [$ARTIFACT_REGISTRY_NO_HTTPS]
      --proxy string             proxy backend registry hostname (and port if not 443 or 80) [$ARTIFACT_REGISTRY_PROXY]
      --proxy-client-ca string   proxy tls client certificate authority [$ARTIFACT_REGISTRY_PROXY_CLIENT_CA]
      --proxy-insecure           disable proxy registry client tls verification [$ARTIFACT_REGISTRY_PROXY_INSECURE]
      --proxy-no-https           disable proxy registry client https [$ARTIFACT_REGISTRY_PROXY_NO_HTTPS]
      --proxy-password string    proxy registry password [$ARTIFACT_REGISTRY_PROXY_PASSWORD]
      --proxy-user string        proxy registry user [$ARTIFACT_REGISTRY_PROXY_USER]
      --tag-artifacts            tag artifacts manifests [$ARTIFACT_REGISTRY_TAG_ARTIFACTS]
      --tls-cert string          tls certificate [$ARTIFACT_REGISTRY_TLS_CERT]
      --tls-key string           tls key [$ARTIFACT_REGISTRY_TLS_KEY]
```

### SEE ALSO

* [lkard completion](lkard_completion.md)	 - Generate the autocompletion script for the specified shell
* [lkard version](lkard_version.md)	 - Print the version informations and exit

