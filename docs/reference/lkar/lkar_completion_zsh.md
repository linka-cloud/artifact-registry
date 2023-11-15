## lkar completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(lkar completion zsh)

To load completions for every new session, execute once:

#### Linux:

	lkar completion zsh > "${fpath[1]}/_lkar"

#### macOS:

	lkar completion zsh > $(brew --prefix)/share/zsh/site-functions/_lkar

You will need to start a new shell for this setup to take effect.


```
lkar completion zsh [flags]
```

### Options

```
  -h, --help              help for zsh
      --no-descriptions   disable completion descriptions
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

* [lkar completion](lkar_completion.md)	 - Generate the autocompletion script for the specified shell

