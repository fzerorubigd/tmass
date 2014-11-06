# Auto-complete for bash and zsh

## Bash

copy bash_tmass in your bash-autocomplete folder or simply add this into your `.bashrc` :

``` .  /path/to/this/file/bash_tmass```

## zsh

Rename the zsh_tmass into _tmass and copy it inside the zsh autocomplete folder like this (in Archlinux) :
```
cp zsh_tmass /usr/share/zsh/site-functions/_tmass
```

or add this into your `.zshrc` :

```
fpath=("/path/to/_tmass" $fpath)
autoload -U compinit
compinit
```
