## lkar completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	lkar completion fish | source

To load completions for every new session, execute once:

	lkar completion fish > ~/.config/fish/completions/lkar.fish

You will need to start a new shell for this setup to take effect.


```
lkar completion fish [flags]
```

### Options

```
  -h, --help              help for fish
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

