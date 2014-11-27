# tmass

Simple and no dependency session manager with load and save ability for [tmux](http://tmux.sourceforge.net/) written in Go (*NOT yet stable.*)

# Why?

I need this to manage my tmux session, mostly in docker container and I need this to be no-dependency. the other I found usable are all depend on Ruby or Python.

The other goal is to automatically create a bash script base on this config. (Not implemented yet.)

## Installation

Install from source using go get :

```go get -u -v github.com/fzerorubigd/tmass```

Or get it from [gobuild.io](http://gobuild.io/github.com/fzerorubigd/tmass)

## Usage

The config file is much like [teamocil](http://www.teamocil.com/ ) but compatibility is not a goal. (currently is different in some ways)

```bash
# Create default config directory
mkdir $HOME/.config/tmass

# 1- Create a tmass config manually :
# Create a sample config
touch $HOME/.config/tmass/sample.yml
# Edit your configuration
$EDITOR $HOME/.config/tmass/sample.yml

# 2- Or Save an existing tmux session
# Run tmux and create your favorite layout inside it
# rename the session if you want to
tmux rename-session sample
# Save the layout with tmass
tmass save sample
# also could user --target to rename saved session name
tmass save --target=sample2 sample

# Run tmass with your configuration for sample (sample and sample.yml are equal)

tmass load sample
```

for better usage help, use ```tmass help```

## Known issue
 - In saving session, the saved command are always without the parameters part. for example `ls -al` is `ls`, this is a tmux limitation
 - The window name in most configs are ignored
 - If the pane count are more than tmux limit, then tmass try to create new window instead of split-window. In this case there is some problem with layout selection.

## Config file

The config file is in yaml format. each file is a session and each session can have multiple window and each window can have any number of pane.

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
| root | The root directory of this pane, the pane switch to this directory before running the commands. default to current dir|
| focus| Is this pane is active?|

## Sample config

Create a two pane window :

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
        root: ~/workspase/
```

A session with two window, the first one contain 3 pane and the next 2 pane :
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

copy bash_tmass in your bash-autocomplete folder or simply add this into your `.bashrc` :

``` .  /path/to/this/file/bash_tmass```

### zsh

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

## Notes/TODO

 * Implement the save to bash script
 * Need more example :)
 * The name is suggested by [@beygoo](https://twitter.com/beygoo) :) ask him why.
