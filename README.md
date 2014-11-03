#tmass

Simple and no dependency session manager for [tmux](http://tmux.sourceforge.net/) written in Go (*NOT yet stable.*)

#Why?

I need this to manage my tmux session, mostly in docker container and I need this to be no-dependency. the other I found usable are all depend on Ruby or Python.

The other goal is to automatically create a bash script base on this config and also save a running tmux session into a layout file.(Not implemented yet.)

##Installation

Install from source using go get :

```go get -u -v github.com/fzerorubigd/tmass```

Or get it from [gobuild.io](http://gobuild.io/github.com/fzerorubigd/tmass)

##Usage

The config file is much like [teamocil](http://www.teamocil.com/ ) but compatibility is not a goal. (currently is different in some ways)

```bash
# Create default config directory
mkdir $HOME/.config/tmass
# Create a sample config
touch $HOME/.config/tmass/sample.yml
# Edit your configuration
$EDITOR $HOME/.config/tmass/sample.yml
# Run tmass with your configuration
tmass sample
```

There is some parameter for tmass :
```
Usage of tmass:
 tmass [OPTIONS] layout
 options are :
  -f, --forcenew=false: Force create new session, default is false if run tmass inside a tmux session, true otherwise.
  -l, --layout-dir="$HOME/.config/tmass/": Layout directory, contain layout files
  -r, --rename=false: Use another name if session name  already exists
```

##Config file

The config file is in yaml format. each file is a session and each session can have multiple window and each window can have any number of pane.

###Session

| Property | Description |
| ---------| ------------|
| name | The name of this session. optional|
| windows| `Array` of `Window` |

###Window

| Property | Description |
| ---------| ------------|
| name | Name of this window. normally ignored by tmux config|
| layout| Layout to use with this window. see the tmux man page about layouts |
| panes | `Array` of `Pane` or a `string`, the command to run inside the pane|

The window name is normally ignored by tmux, since any command inside the window can change the window name. If you want to force tmux to use only your name then disable tmux `allow-rename` option.

###Pane

| Property | Description |
| ---------| ------------|
| commands | `Array` of `string`, each item is a command to be executed inside the pane |
| root | The root directory of this pane, the pane switch to this directory before running the commands. default to current dir|
| focus| Is this pane is active? Not yet implemented|

##Sample config

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

 * Need more example :)
 * The name is suggested by [@beygoo](https://twitter.com/beygoo) :) ask him why.
