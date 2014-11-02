package main

import (
	"fmt"
	"os/exec"
	"strings"

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

	flag.Parse()

	fmt.Println(*forceNew, *layoutDir)

	sess, err := LoadSessionFromFile("/home/f0rud/.teamocil/sample.yml")
	if err != nil {
		panic(err)
	}

	sess.ForceNew = *forceNew

	for _, k := range BuildSession(sess) {
		fmt.Println("tmux " + k)
		err := exec.Command("/usr/bin/tmux", strings.Split(k, " ")...).Run()
		if err != nil {
			panic(err)
		}
	}
}
