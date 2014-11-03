package main

import (
	"fmt"
	"log"

	flag "github.com/ogier/pflag"
)

func main() {
	forceNew := flag.BoolP(
		"forcenew",
		"f",
		!IsInsideTmux(),
		`Force create new session, default is false if run tmass inside a tmux session, true otherwise.`,
	)

	home, err := GetHomeDir()
	if err != nil {
		panic(err)
	}

	layoutDir := flag.StringP(
		"layout",
		"l",
		home+"/.config/tmass/",
		`Layout directory, contain layout files`,
	)

	rename := flag.BoolP(
		"rename",
		"r",
		false,
		`Use another name if session name exists already`,
	)

	flag.Parse()

	fmt.Println(*forceNew, *layoutDir)

	sess, err := LoadSessionFromFile("/home/f0rud/.teamocil/sample.yml")
	if err != nil {
		panic(err)
	}

	sess.ForceNew = *forceNew

	if err := sess.BuildSession("tmux", *rename); err != nil {
		log.Fatal(err)
	}
}
