# tmass

Simple and no-dependency session manager with load and save ability for [tmux](http://tmux.sourceforge.net/) written in Go (*NOT yet stable*).

# Why?

I need this to manage my tmux sessions, mostly in docker containers and I need this to be no-dependency. The others I found usable, all depend on Ruby or Python.

The other goal is to automatically create a bash script based on the config (not implemented yet).

## Installation

Clone the source and in source folder run `make all`

## Usage

The config file is much like [teamocil](http://www.teamocil.com/), but compatibility is not a goal (currently it is different in some ways).

```bash
# Create default config directory
mkdir $HOME/.config/tmass

# 1- Create a tmass config manually:
# Create a sample config
touch $HOME/.config/tmass/sample.yml
# Edit your configuration
$EDITOR $HOME/.config/tmass/sample.yml

# 2- Or Save an existing tmux session
# Run tmux and create your favorite layout inside it
# Rename the session if you want to
tmux rename-session sample
# Save the layout with tmass
tmass save sample
# also --target can be used to rename a saved session
tmass save --target=sample2 sample

# Run tmass with your sample configuration (sample and sample.yml are equal)

tmass load sample
```

for better usage help, use ```tmass help```

## Known issues
 - When saving a session, the saved commands are always without the parameters part. For example `ls -al` is `ls`, this is a tmux limitation.
 - The window name in most configs are ignored.
 - If the pane count are more than tmux limit, then tmass tries to create a new window instead of split-window. In this case there are some problems with layout selection.

## Config file

The config file is in yaml format. Each file is a session and each session can have multiple windows and each window can have any number of panes.

### Session

| Property | Description |
| ---------| ------------|
| name | The name of this session. optional|
| windows| `Array` of `Window` |

### Window

| Property | Description |
| ---------| ------------|
| name | Name of this window. normally ignored by tmux config|
| layout| Layout to use with this window. see the tmux man page about layouts |
| panes | `Array` of `Pane` or `string`, the command to run inside the pane|

The window name is normally ignored by tmux, since any command inside the window can change the window name. If you want to force tmux to use only your name then disable tmux `allow-rename` option.

### Pane

| Property | Description |
| ---------| ------------|
| commands | `Array` of `string`, each item is a command to be executed inside the pane |
| root | The root directory of this pane, the pane switches to this directory before running the commands. default to current dir|
| focus| Is this pane active?|

## Sample config

Create a two pane window:

```yml
name: two-pane-session
windows:
  - name: normally-ignored
    layout: even-horizontal
    panes:
      - git status
      - commands:
          - ls
          - echo "Hey"
        root: ~/workspace/
```

A session with two windows, the first one contains 3 panes and the next one 2 panes:
```yml
name: two-window-per-session
windows:
  - name: window-1
    layout: main-vertical
    panes:
      - ls -al
      - #Nothing
      - echo its me
  - name: window-2
    layout: even-vertical
    panes:
      - cat /etc/resolv.conf
      - #ignored
```

## Auto complete for bash and zsh

### Bash

Copy bash_tmass in your bash-autocomplete folder or simply add this into your `.bashrc` :

``` .  /path/to/this/file/bash_tmass```

### zsh

Rename the zsh_tmass to _tmass and copy it inside the zsh autocomplete folder like this (in Archlinux) :
```
cp zsh_tmass /usr/share/zsh/site-functions/_tmass
```

or add this into your `.zshrc` :

```
fpath=("/path/to/_tmass" $fpath)
autoload -U compinit
compinit
```

## Notes/TODO

 * Implement the save to bash script
 * Need more example :)
 * The name is suggested by [@beygoo](https://twitter.com/beygoo) :) ask him why.
