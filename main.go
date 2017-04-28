package main

import (
	"log"
	"os"
	"path"
	"strings"

	"github.com/fzerorubigd/tmass/tmux"
	"github.com/fzerorubigd/tmass/version"
	"github.com/juliengk/go-utils/user"
	"github.com/mitchellh/colorstring"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func main() {
	var (
		tmuxCmd   string
		tmuxArgs  string
		layoutDir string
		target    string
		attach    bool
	)

	u := user.New()

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
			if layoutDir[len(layoutDir)-1:] != "/" {
				layoutDir += "/"
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

			var tArgs []string
			if tmuxArgs != "" {
				tArgs = strings.Split(tmuxArgs, " ")
			} else {
				tArgs = make([]string, 0)
			}

			if target != "" {
				sess.Name = target
			}

			if err := tmux.BuildSession(sess, tmuxCmd, tArgs, attach); err != nil {
				log.Fatal(err)
			}
			log.Print(colorstring.Color("[green]Session has been loaded"))
		},
	}

	load.Flags().StringVar(
		&target,
		"target",
		"",
		`try to load session using this name, empth name is default, anything else is the session name`,
	)

	load.Flags().BoolVarP(
		&attach,
		"attach",
		"a",
		false,
		`Attach to sessio if its exists already, if not exists then create it`,
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
			if layoutDir[len(layoutDir)-1:] != "/" {
				layoutDir += "/"
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

	version := &cobra.Command{
		Use:   "version",
		Short: `Show the tmass version information`,
		Long:  `All software has versions. This is tmass's`,
		Run: func(cmd *cobra.Command, args []string) {
			info := version.New()
			info.ShowVersion()
		},
	}

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
		u.HomeDir+"/.config/tmass/",
		`Layout directory, contain layout files`,
	)

	// A dirty hack for better document
	load.SetUsageTemplate(strings.Replace(root.UsageTemplate(), "[flags]", "[flags] layoutname", -1))
	save.SetUsageTemplate(strings.Replace(root.UsageTemplate(), "[flags]", "[flags] sessiontname", -1))
	root.AddCommand(load, save, version)
	root.Execute()
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
