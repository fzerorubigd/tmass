package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mitchellh/colorstring"
	flag "github.com/ogier/pflag"
	"gopkg.in/yaml.v2"
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n %s [OPTIONS] layout\n options are : \n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
}

func main() {
	tmux := flag.String(
		"tmux",
		"tmux",
		`The tmux command to use, just if tmux is not in the $PATH`,
	)

	forceNew := flag.BoolP(
		"forcenew",
		"f",
		!IsInsideTmux(),
		`Force create new session, default is false if run tmass inside a tmux session, true otherwise.`,
	)

	home, err := GetHomeDir()
	if err != nil {
		log.Panic(err)
	}

	layoutDir := flag.StringP(
		"layout-dir",
		"l",
		home+"/.config/tmass/",
		`Layout directory, contain layout files`,
	)

	rename := flag.BoolP(
		"rename",
		"r",
		false,
		`Use another name if session name already exists`,
	)

	save := flag.Bool(
		"save",
		false,
		`Try to save the session from tmux, if this switch is used all other switchs are ignored (except for --tmux) and the layout must be exist as a session in a running instanse of tmux`,
	)

	flag.Parse()
	if len(os.Args)-2 != flag.NFlag() {
		log.Println("wrong number of argument")
		flag.Usage()
		return
	}

	// The last arg is the layout name
	layout := os.Args[len(os.Args)-1]

	filename := *layoutDir + layout + ".yml"
	notExists := false
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		notExists = true
	}

	if *save {
		if !notExists {
			log.Fatalf("file already exists: %s", filename)
		}

		s, err := LoadSessionFromTmux(*tmux, layout)
		if err != nil {
			log.Fatal(err)
		}

		o, err := yaml.Marshal(*s)
		if err != nil {
			log.Fatal(err)
		}

		if err := SaveSessionToFile(o, filename); err != nil {
			log.Fatal(err)
		}
		log.Printf(colorstring.Color("[green]The file %s has been written, PLEASE verify that, the name and commands part mostly are not correct. see Known issue in readme."), filename)
	} else {
		if notExists {
			log.Fatalf("no such file: %s", filename)
		}

		sess, err := LoadSessionFromFile(filename)
		if err != nil {
			log.Fatal(err)
		}

		sess.ForceNew = *forceNew

		if err := sess.BuildSession(*tmux, *rename); err != nil {
			log.Fatal(err)
		}
		log.Print(colorstring.Color("[green]Session has been loaded"))
	}
}
