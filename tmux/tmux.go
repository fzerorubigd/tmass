package tmux

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"os/user"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/mitchellh/colorstring"
	"gopkg.in/yaml.v2"
)

type Session struct {
	// This is set from command line.
	ForceNew bool     `yaml:"-"`
	Name     string   `yaml:"name,omitempty"`
	Windows  []Window `yaml:"windows"`
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

type Command struct {
	Parts []string
}

var (
	IgnoredCmd []string
	DefaultCmd string
)

func (m *Command) Add(part ...string) {
	if m.Parts == nil {
		m.Parts = make([]string, 0)
	}
	m.Parts = append(m.Parts, part...)
}

func (m *Command) String() string {
	return strings.Join(m.Parts, " ")
}

func (m *Command) Execute(base string) (string, error) {
	//fmt.Println(m.String())
	out, err := exec.Command(base, m.Parts...).Output()

	if err != nil {
		err = fmt.Errorf("failed to execute %s %s : %s \n %s", base, m.String(), err.Error(), string(out))
	}
	return strings.TrimSpace(string(out)), err
}

func (m *Command) Clear() {
	m.Parts = nil
}

func BuildSession(s *Session, tmux string, rename bool) error {
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

	cmd := Command{}
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
			// tmux -P switch for new window return data for window not the first pane
			//TODO: Default is zero, if default is changed by user?
			s.Windows[i].RealPane[0].identifier = s.Name + ":0.0"
		} else {
			// If this is a rename session Command
			if i == 0 {
				if _, err := cmd.Execute(tmux); err != nil {
					return err
				}
			}
			c := Command{}
			c.Add("new-window", "-P", "-t", s.Name, "-n", s.Windows[i].Name, "-c", s.Windows[i].RealPane[0].Root)
			if n, err := c.Execute(tmux); err != nil {
				return err
			} else {
				s.Windows[i].RealPane[0].identifier = n
			}
		}
		cf, err := BuildPane(&s.Windows[i], tmux, s)
		if err != nil {
			return err
		}
		// The problem is, layout may contain the focus. so for setting focus, we need to call it after setting layout
		c := Command{[]string{"select-layout", s.Windows[i].Layout}}
		if _, err := c.Execute(tmux); err != nil {
			return err
		}

		if cf != nil {
			if _, err := cf.Execute(tmux); err != nil {
				return err
			}
		}
	}

	return nil
}

func BuildPane(w *Window, tmux string, s *Session) (*Command, error) {

	cf := Command{}
	for i, p := range w.RealPane {
		if i > 0 { // The first pane is created when the window is created
			c0 := Command{[]string{"split-window", "-P", "-c", p.Root}}
			if n, err := c0.Execute(tmux); err != nil {
				return nil, err
			} else {
				p.identifier = n
			}

		}
		c1 := Command{[]string{"send-keys", "-t", p.identifier, strings.Join(p.Commands, ";")}}
		c2 := Command{[]string{"send-keys", "-t", p.identifier, "Enter"}}
		if _, err := c1.Execute(tmux); err != nil {
			return nil, err
		}
		if _, err := c2.Execute(tmux); err != nil {
			return nil, err
		}
		if p.Focus {
			cf.Clear()
			cf.Add("select-pane", "-t", p.identifier)
		}
	}

	if cf.String() != "" {
		return &cf, nil
	}
	return nil, nil
}

func LoadSessionFromTmux(tmux, session string) (*Session, error) {
	sess := Session{Name: session}
	sess.Windows = make([]Window, 0)
	cmd := Command{}
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
	cmd := Command{}
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
			for _, v := range IgnoredCmd {
				if v == parts[1] {
					parts[1] = DefaultCmd
					break
				}
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

func LoadSession(data []byte) (*Session, error) {
	session := &Session{}
	if err := yaml.Unmarshal(data, session); err != nil {
		return nil, err
	}

	appRoot, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	home, err := getHomeDir()
	if err != nil {
		return nil, err
	}

	for i := range session.Windows {
		session.Windows[i].RealPane = make([]Pane, 0)
		for _, iv := range session.Windows[i].Panes {
			switch iv.(type) {
			case string:
				session.Windows[i].RealPane = append(session.Windows[i].RealPane, Pane{Commands: []string{iv.(string)}, Root: appRoot})
			default:
				// Anything else? we can accept pane, lets use yaml again!
				y, err := yaml.Marshal(iv)
				if err != nil {
					return nil, err
				}
				p := &Pane{}
				if err := yaml.Unmarshal(y, p); err != nil {
					return nil, err
				}
				if p.Root == "" {
					p.Root = appRoot // Make sure each pane has a root
				} else if p.Root[:1] == "~" {
					p.Root = home + p.Root[1:]
				}
				session.Windows[i].RealPane = append(session.Windows[i].RealPane, *p)
			}
		}

		if len(session.Windows[i].RealPane) == 0 {
			return nil, errors.New("each window must have at least one pane")
		}
	}

	if len(session.Windows) == 0 {
		return nil, errors.New("each session must have at least one window")
	}

	return session, nil
}

func LoadSessionFromFile(fileName string) (*Session, error) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return LoadSession(content)
}

func SaveSessionToFile(sess []byte, fileName string) error {
	return ioutil.WriteFile(fileName, sess, 0644)
}

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

func init() {
	DefaultCmd = `echo "Default CMD"`
	IgnoredCmd = []string{os.Args[0]}
	if s := os.Getenv("SHELL"); s != "" {
		b := path.Base(s)
		IgnoredCmd = append(IgnoredCmd, s, b)
	}
}
