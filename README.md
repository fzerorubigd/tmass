#tmass

Simple and no dependency session manager with load and save ability for [tmux](http://tmux.sourceforge.net/) written in Go (*NOT yet stable.*)

#Why?

I need this to manage my tmux session, mostly in docker container and I need this to be no-dependency. the other I found usable are all depend on Ruby or Python.

The other goal is to automatically create a bash script base on this config. (Not implemented yet.)

##Installation

Install from source using go get :

```go get -u -v github.com/fzerorubigd/tmass```

Or get it from [gobuild.io](http://gobuild.io/github.com/fzerorubigd/tmass)

##Usage

The config file is much like [teamocil](http://www.teamocil.com/ ) but compatibility is not a goal. (currently is different in some ways)

```bash
# Create default config directory
mkdir $HOME/.config/tmass

# 1- Create a tmass config manually : 
# Create a sample config
touch $HOME/.config/tmass/sample.yml
# Edit your configuration
$EDITOR $HOME/.config/tmass/sample.yml
# Run tmass with your configuration

# 2- Or Save an existing tmux session
# Run tmux and create your favorite layout inside it 
# rename the session if you want to 
tmux rename-session sample
# Save the layout with tmass 
tmass --save sample

# Load the sample session  
tmass sample
```

There is some parameter for tmass :
```
Usage of tmass:
 tmass [OPTIONS] layout
 options are : 
  -f, --forcenew=true: Force create new session, default is false if run tmass inside a tmux session, true otherwise.
  -l, --layout-dir="$HOME/.config/tmass/": Layout directory, contain layout files
  -r, --rename=false: Use another name if session name already exists
      --save=false: Try to save the session from tmux, if this switch is used all other switchs are ignored (except for --tmux) and the layout must be exist as a session in a running instanse of tmux
      --tmux="tmux": The tmux command to use, just if tmux is not in the $PATH

layout normally is the layout to load, with --save, layout is the file name to create.
```

## Known issue 
 - In saving session, the saved command are always without the parameters part. for example `ls -al` is `ls`, this is a tmux limitation
 - The window name in most configs are ignored 

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
| panes | `Array` of `Pane` or `string`, the command to run inside the pane|

The window name is normally ignored by tmux, since any command inside the window can change the window name. If you want to force tmux to use only your name then disable tmux `allow-rename` option.

###Pane

| Property | Description |
| ---------| ------------|
| commands | `Array` of `string`, each item is a command to be executed inside the pane |
| root | The root directory of this pane, the pane switch to this directory before running the commands. default to current dir|
| focus| Is this pane is active?|

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

 * Implement the save to bash script 
 * Need more example :)
 * The name is suggested by [@beygoo](https://twitter.com/beygoo) :) ask him why.
