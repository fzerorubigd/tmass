package main

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
)

func BuildSession(sess *Session) []string {

	if sess.Name == "" {
		sess.Name = "teco-session-" + strconv.Itoa(rand.Int())
	}

	if IsSessionExists(sess.Name) {
		panic("session already exists") // TODO : clean this
	}

	// If there is a force new switch, then forcing create new tmux session, else rename current
	firstCmd := ""
	if sess.ForceNew {
		firstCmd = "new-session -d "
		firstCmd += "-s "
	} else {
		firstCmd = "rename-session "
	}
	firstCmd += sess.Name

	result := make([]string, 0)

	// Normally numbers are 0 base unless there is a change in tmux config
	counter := 0

	for i := range sess.Windows {
		if sess.ForceNew && counter == 0 {
			firstCmd += " -n " + sess.Windows[i].Name + " -c " + sess.Windows[i].RealPane[0].Root
		} else {
			cmd := "new-window -t " + sess.Name + " -n " + sess.Windows[i].Name + " -c " + sess.Windows[i].RealPane[0].Root
			result = append(result, cmd)
		}

		result = append(result, BuildPane(sess, &sess.Windows[i], counter)...)
		counter += 1
	}

	return append([]string{firstCmd}, result...)
}

func BuildPane(sess *Session, win *Window, index int) []string {
	result := make([]string, 0)

	counter := 0
	//chosen := ""
	for _, p := range win.RealPane {
		cmd := []string{
			"send-keys -t " + sess.Name + ":" + strconv.Itoa(index) + " " + strings.Join(p.Commands, ";"),
			"send-keys -t " + sess.Name + ":" + strconv.Itoa(index) + " Enter",
		}
		if counter != 0 {
			result = append(result, "split-window -c "+p.Root)
		}
		result = append(result, cmd...)
		//if p.Focus {
		//	chosen = "select-pane -t "
		//}

		counter += 1
	}

	// Now should add the layout stuff
	result = append(result, "select-layout "+win.Layout)

	return result
}

func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

//TODO : write this
func IsSessionExists(name string) bool {
	//out, _ := exec.Command("tmux", "ls").Output()

	return false
}
