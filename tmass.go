package main

import (
	"log"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/fzerorubigd/tmass/tmux"
	"github.com/mitchellh/colorstring"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		tmuxCmd   string
		tmuxArgs  string
		forceNew  bool
		layoutDir string
		rename    bool
		target    string
	)

	home, err := getHomeDir()
	if err != nil {
		log.Panic(err)
	}

	root := &cobra.Command{
		Use:   "tmass",
		Short: "tmux session manager with no-dependency",
		Long:  `Manage tmux session, load session from file and save active session into a file`,
	}

	load := &cobra.Command{
		Use:   "load",
		Short: `Load session from file`,
		Long: `Load a session from file and apply it to tmux. the default location is $HOME/.config/tmass
if run this inside tmux session then this apply to current session. if not, this create new session.
use --forcenew to overwrite this`,
		Run: func(cmd *cobra.Command, args []string) {
			// The last arg is the layout name
			layout := strings.Join(args, " ")
			if layout == "" {
				log.Fatalf("the session name is empty")
			}

			filename := layoutDir + layout
			if path.Ext(filename) != ".yml" {
				filename += ".yml"
			}
			checkLayoutDir(layoutDir)
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				log.Fatalf("no such file: %s", filename)
			}

			sess, err := tmux.LoadSessionFromFile(filename)
			if err != nil {
				log.Fatal(err)
			}

			sess.ForceNew = forceNew

			var tArgs []string
			if tmuxArgs != "" {
				tArgs = strings.Split(tmuxArgs, " ")
			} else {
				tArgs = make([]string, 0)
			}

			if err := tmux.BuildSession(sess, tmuxCmd, tArgs, rename); err != nil {
				log.Fatal(err)
			}
			log.Print(colorstring.Color("[green]Session has been loaded"))
		},
	}

	load.Flags().BoolVarP(
		&forceNew,
		"forcenew",
		"f",
		!tmux.IsInsideTmux(),
		`Force create new session, default is false if run tmass inside a tmux session, true otherwise.`,
	)

	load.Flags().BoolVarP(
		&rename,
		"rename",
		"r",
		false,
		`Use another name if session name already exists`,
	)

	save := &cobra.Command{
		Use:   "save",
		Short: `Save session into file`,
		Long:  `Save a session into a file. the file name is the session name but you can change it.`,
		Run: func(cmd *cobra.Command, args []string) {
			// The last arg is the layout name
			layout := strings.Join(args, " ")
			if layout == "" {
				log.Fatalf("the session name is empty")
			}
			if target == "" {
				target = layout
			}
			filename := layoutDir + target
			if path.Ext(filename) != ".yml" {
				filename += ".yml"
			}
			checkLayoutDir(layoutDir)
			if _, err := os.Stat(filename); !os.IsNotExist(err) {
				log.Fatalf("file already exists: %s", filename)
			}

			var tArgs []string
			if tmuxArgs != "" {
				tArgs = strings.Split(tmuxArgs, " ")
			} else {
				tArgs = make([]string, 0)
			}

			s, err := tmux.LoadSessionFromTmux(tmuxCmd, tArgs, layout)
			if err != nil {
				log.Fatal(err)
			}

			o, err := yaml.Marshal(*s)
			if err != nil {
				log.Fatal(err)
			}

			if err := tmux.SaveSessionToFile(o, filename); err != nil {
				log.Fatal(err)
			}
			log.Printf(colorstring.Color("[green]The file %s has been written, PLEASE verify that, the name and commands part mostly are not correct. see Known issue in readme."), filename)
		},
	}

	save.Flags().StringVar(
		&target,
		"target",
		"",
		`Change the save target, default is session name`,
	)

	root.PersistentFlags().StringVar(
		&tmuxCmd,
		"tmux",
		"tmux",
		`The tmux command to use, just if tmux is not in the $PATH`,
	)

	root.PersistentFlags().StringVar(
		&tmuxArgs,
		"tmux-args",
		"",
		`Extra arguments to use with tmux`,
	)

	root.PersistentFlags().StringVarP(
		&layoutDir,
		"layout-dir",
		"l",
		home+"/.config/tmass/",
		`Layout directory, contain layout files`,
	)

	// A dirty hack for better document
	load.SetUsageTemplate(strings.Replace(root.UsageTemplate(), "[flags]", "[flags] layoutname", -1))
	save.SetUsageTemplate(strings.Replace(root.UsageTemplate(), "[flags]", "[flags] sessiontname", -1))
	root.AddCommand(load, save)
	root.Execute()
}

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func checkLayoutDir(ld string) {

	// check if the source dir exist
	src, err := os.Stat(ld)
	if err != nil {
		log.Fatal(err)
	}

	// check if the source is indeed a directory or not
	if !src.IsDir() {
		log.Fatal("Source is not a directory, Please create this directory or pass --layout-dir")
	}
}
