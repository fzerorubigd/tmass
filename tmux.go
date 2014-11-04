package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/mitchellh/colorstring"
)

type Session struct {
	// This is set from command line.
	ForceNew bool     `yaml:"-"`
	Name     string   `yaml:"name,omitempty"`
	Windows  []Window `yaml:"windows"`
	// In detach mode, we need to keep track of created windows
	counter int `yaml:"-"`
}

type Window struct {
	Name string `yaml:"name"`
	// This is some kind of incompatibility with teamocil, But I need to each pane have its own start path
	// The value from config is ignored here
	Root     string        `yaml:"root,omitempty"`
	Layout   string        `yaml:"layout"`
	Panes    []interface{} `yaml:"panes"`
	RealPane []Pane        `yaml:"-"`
}

type Pane struct {
	Commands   []string `yaml:"commands"`
	Focus      bool     `yaml:"focus,omitempty"`
	Root       string   `yaml:"root,omitempty"`
	identifier string   `yaml:"-"`
}

type command struct {
	Parts []string
}

func (m *command) Add(part ...string) {
	if m.Parts == nil {
		m.Parts = make([]string, 0)
	}
	m.Parts = append(m.Parts, part...)
}

func (m *command) String() string {
	return strings.Join(m.Parts, " ")
}

func (m *command) Execute(base string) (string, error) {
	//fmt.Println(m.String())
	out, err := exec.Command(base, m.Parts...).Output()

	return strings.TrimSpace(string(out)), err
}

func (s *Session) BuildSession(tmux string, rename bool) error {
	if s.Name == "" {
		s.Name = "tmass-session-" + strconv.Itoa(rand.Int())
	}

	// Wow this code is creepy :/
	for IsSessionExists(s.Name) {
		if rename {
			s.Name = "tmass-session-" + strconv.Itoa(rand.Int())
		} else {
			return fmt.Errorf("session with name %s already exists", s.Name)
		}
	}

	cmd := command{}
	if s.ForceNew {
		cmd.Add("new-session", "-d", "-s")
	} else {
		cmd.Add("rename-session")
	}
	cmd.Add(s.Name)

	for i := range s.Windows {
		if s.ForceNew && i == 0 { // First window is created when new session is started
			cmd.Add("-n", s.Windows[i].Name, "-c", s.Windows[i].RealPane[0].Root)
			if _, err := cmd.Execute(tmux); err != nil {
				return err
			}
			s.Windows[i].RealPane[0].identifier = s.Name + ":0" //TODO: Default is zero, if default is changed by user?
		} else {
			// If this is a rename session command
			if i == 0 {
				if _, err := cmd.Execute(tmux); err != nil {
					return err
				}
			}
			c := command{}
			c.Add("new-window", "-P", "-t", s.Name, "-n", s.Windows[i].Name, "-c", s.Windows[i].RealPane[0].Root)
			if n, err := c.Execute(tmux); err != nil {
				return err
			} else {
				s.Windows[i].RealPane[0].identifier = n
			}
		}
		if err := s.Windows[i].BuildPane(tmux, s); err != nil {
			return err
		}

		c := command{[]string{"select-layout", s.Windows[i].Layout}}
		if _, err := c.Execute(tmux); err != nil {
			return err
		}
	}

	return nil
}

func (w *Window) BuildPane(tmux string, s *Session) error {
	// TODO : Support initial focus
	for i, p := range w.RealPane {
		if i > 0 { // The first pane is created when the window is created
			c0 := command{[]string{"split-window", "-P", "-c", p.Root}}
			if n, err := c0.Execute(tmux); err != nil {
				return err
			} else {
				p.identifier = n
			}

		}
		c1 := command{[]string{"send-keys", "-t", p.identifier, strings.Join(p.Commands, ";")}}
		c2 := command{[]string{"send-keys", "-t", p.identifier, "Enter"}}
		if _, err := c1.Execute(tmux); err != nil {
			return err
		}
		if _, err := c2.Execute(tmux); err != nil {
			return err
		}
	}

	return nil
}

func LoadSessionFromTmux(tmux, session string) (*Session, error) {
	sess := Session{Name: session}
	sess.Windows = make([]Window, 0)
	//tmux list-windows -t mine -F '#S:#I|#{window_panes}|#{window_layout}'
	cmd := command{}
	cmd.Add("list-window", "-t", session, "-F", "#S:#I|#{window_name}|#{window_layout}")
	if out, err := cmd.Execute(tmux); err != nil {
		return nil, err
	} else {
		for _, s := range strings.Split(out, "\n") {
			parts := strings.Split(s, "|")
			if len(parts) != 3 {
				log.Println(colorstring.Color("[red][_yellow_]Invalid count! ignoring this window!"))
				continue
			}

			if w, err := LoadWindowFromTmux(tmux, parts[0], parts[1], parts[2]); err != nil {
				return nil, err
			} else {
				sess.Windows = append(sess.Windows, *w)
			}
		}
	}

	return &sess, nil

}

func LoadWindowFromTmux(tmux, window, name, layout string) (*Window, error) {
	// The real pane is not used here. ignore it
	w := Window{Name: name, Layout: layout, Panes: make([]interface{}, 0)}
	cmd := command{}
	//tmux list-panes -t mine:1 -F '#P|#{pane_current_path}|#{pane_current_command}'
	cmd.Add("list-pane", "-t", window, "-F", "#{pane_current_path}|#{pane_current_command}|#{pane_active}")
	if out, err := cmd.Execute(tmux); err != nil {
		return nil, err
	} else {
		for _, s := range strings.Split(out, "\n") {
			parts := strings.Split(s, "|")
			if len(parts) != 3 {
				log.Println(colorstring.Color("[red][_yellow_]Invalid count! ignoring this pane!"))
				continue
			}
			p := Pane{Commands: []string{parts[1]}, Root: parts[0], Focus: parts[2] == "1"}
			w.Panes = append(w.Panes, p)
		}
	}

	return &w, nil
}

func IsInsideTmux() bool {
	// Simply, if the TMUX is set in env, We are in it :)
	return os.Getenv("TMUX") != ""
}

func IsSessionExists(name string) bool {
	out, err := exec.Command("tmux", "ls").Output()
	if err != nil {
		// This means either tmux is not running at all, or any other bad thing.
		// Just return false here
		return false
	}
	re := regexp.MustCompile("(?m)^([^: \t\n]*):")
	matches := re.FindAllStringSubmatch(string(out), -1)
	for _, m := range matches {
		// m is atleast in size 2, I have a subpattern
		if m[1] == name {
			return true
		}
	}

	return false
}
