package tmux

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/juliengk/go-utils/user"
	"github.com/mitchellh/colorstring"
	"gopkg.in/yaml.v2"
)

// Session handle a tmux session, each session contain many Window
type Session struct {
	// This is set from command line.
	Attach  bool     `yaml:"-"`
	Name    string   `yaml:"name,omitempty"`
	Windows []Window `yaml:"windows"`
}

// Window handle tmux window, each window can have multiple pane.
// tmux has limit for pane count base window layout
type Window struct {
	Name string `yaml:"name"`
	// This is some kind of incompatibility with teamocil, But I need to each pane have its own start path
	// The value from config is ignored here
	Root     string        `yaml:"root,omitempty"`
	Layout   string        `yaml:"layout"`
	Panes    []interface{} `yaml:"panes"`
	RealPane []Pane        `yaml:"-"`
}

// Pane handle each pane (single command line) in tmux
type Pane struct {
	Commands   []string `yaml:"commands"`
	Focus      bool     `yaml:"focus,omitempty"`
	Root       string   `yaml:"root,omitempty"`
	identifier string   `yaml:"-"`
}

// Command is a helper for executable command inside tmux pane
type Command struct {
	Parts []string
}

var (
	// IgnoredCmd is list of commands ignored by save session
	IgnoredCmd []string
	// DefaultCmd is a command used when the command is ignored
	DefaultCmd string
	// Copy of env whitout tmux related env
	tmuxENV []string
)

// Add a part to command
func (m *Command) Add(part ...string) {
	if m.Parts == nil {
		m.Parts = make([]string, 0)
	}
	m.Parts = append(m.Parts, part...)
}

// Convert command to string
func (m *Command) String() string {
	return strings.Join(m.Parts, " ")
}

// Execute this command
func (m *Command) Execute(base string, args []string) (string, error) {
	args = append(args, m.Parts...)
	cmd := exec.Command(base, args...)
	cmd.Env = tmuxENV

	out, err := cmd.CombinedOutput()

	if err != nil {
		err = fmt.Errorf("failed to execute %s %s : %s \n %s", base, strings.Join(args, " "), err.Error(), string(out))
	}
	//fmt.Println(m.String(), "==>", string(out))
	return strings.TrimSpace(string(out)), err
}

// Clear the current command
func (m *Command) Clear() {
	m.Parts = nil
}

// BuildSession build a session based on Session structure
func BuildSession(s *Session, tmux string, args []string, attach bool) error {

	if s.Name == "" {
		s.Name = "tmass-session-" + strconv.Itoa(rand.Int())
	}
	// Wow this code is creepy :/
	if IsSessionExists(s.Name) {
		if !attach {
			return fmt.Errorf("session with name %s already exists, use the --attach switch to attach to it or use --target for overwrite name", s.Name)
		}
		s.Attach = true
	}

	for i := range s.Windows {
		if !s.Attach && i == 0 { // First window is created when new session is started, if its not an attach session
			cmd := Command{}
			if !s.Attach {
				cmd.Add("new-session", "-d", "-s")
			}
			cmd.Add(s.Name)
			cmd.Add("-n", s.Windows[i].Name, "-c", s.Windows[i].RealPane[0].Root)
			if _, err := cmd.Execute(tmux, args); err != nil {
				return err
			}
			// tmux -P switch for new window return data for window not the first pane
			//TODO: Default is zero, if default is changed by user?
			s.Windows[i].RealPane[0].identifier = s.Name + ":0.0"
		} else {
			c := Command{}
			c.Add("new-window", "-P", "-t", s.Name, "-n", s.Windows[i].Name, "-c", s.Windows[i].RealPane[0].Root)
			n, err := c.Execute(tmux, args)
			if err != nil {
				return err
			}
			s.Windows[i].RealPane[0].identifier = n
		}
		cf, err := BuildPane(&s.Windows[i], tmux, args, s)
		if err != nil {
			return err
		}
		// The problem is, layout may contain the focus. so for setting focus, we need to call it after setting layout
		c := Command{[]string{"select-layout", s.Windows[i].Layout}}
		if _, err := c.Execute(tmux, args); err != nil {
			log.Println(colorstring.Color("[yellow] Failed to apply layout. ignored"))
		}

		if cf != nil {
			if _, err := cf.Execute(tmux, args); err != nil {
				log.Println(colorstring.Color("[yellow] Failed to select pane. ignored"))
			}
		}
	}

	return nil
}

// A simple fallback, Sometime the split-window fails with `create pane failed: pane too small`
// In that case, print a warning and then try to use new-window
func newWindowFallback(w *Window, tmux string, args []string, s *Session, p *Pane) (string, error) {
	log.Println(colorstring.Color("[yellow] Failed to split window. try to create new window."))
	// First try to set layout for old window
	c := Command{[]string{"select-layout", w.Layout}}
	if _, err := c.Execute(tmux, args); err != nil {
		log.Println(colorstring.Color("[yellow] Failed to apply layout. ignored"))
	}
	c.Clear()
	c.Add("new-window", "-P", "-t", s.Name, "-n", w.Name, "-c", p.Root)
	return c.Execute(tmux, args)
}

// BuildPane build a pane from a window
func BuildPane(w *Window, tmux string, args []string, s *Session) (*Command, error) {

	cf := Command{}
	var basePane string
	for i, p := range w.RealPane {
		c := Command{[]string{"select-layout", w.Layout}}
		if _, err := c.Execute(tmux, args); err != nil {
			log.Println(colorstring.Color("[yellow] Failed to apply layout. ignored"))
		}

		if i == 0 {
			basePane = p.identifier // I don't know a good way to find the bigger pane, so just split the first pane
			// TODO do I need to re-apply the layout each time???
		} else { // The first pane is created when the window is created
			c0 := Command{[]string{"split-window", "-P", "-c", p.Root, "-t", basePane}}
			n, err := c0.Execute(tmux, args)
			if err != nil {
				if n, err = newWindowFallback(w, tmux, args, s, &p); err != nil {
					return nil, err
				}
			}
			p.identifier = n
		}
		c1 := Command{[]string{"send-keys", "-t", p.identifier, strings.Join(p.Commands, ";")}}
		c2 := Command{[]string{"send-keys", "-t", p.identifier, "Enter"}}
		if _, err := c1.Execute(tmux, args); err != nil {
			return nil, err
		}
		if _, err := c2.Execute(tmux, args); err != nil {
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

// LoadSessionFromTmux fill session structure from a running instance of tmux
func LoadSessionFromTmux(tmux string, args []string, session string) (*Session, error) {
	sess := Session{Name: session}
	sess.Windows = make([]Window, 0)
	cmd := Command{}
	cmd.Add("list-window", "-t", session, "-F", "#S:#I|#{window_name}|#{window_layout}")
	out, err := cmd.Execute(tmux, args)
	if err != nil {
		return nil, err
	}
	for _, s := range strings.Split(out, "\n") {
		parts := strings.Split(s, "|")
		if len(parts) != 3 {
			log.Println(colorstring.Color("[red][_yellow_]Invalid count! ignoring this window!"))
			continue
		}
		w, err := LoadWindowFromTmux(tmux, args, parts[0], parts[1], parts[2])
		if err != nil {
			return nil, err
		}
		sess.Windows = append(sess.Windows, *w)
	}

	return &sess, nil

}

// LoadWindowFromTmux loads window from a tmux session
func LoadWindowFromTmux(tmux string, args []string, window, name, layout string) (*Window, error) {
	// The real pane is not used here. ignore it
	w := Window{Name: name, Layout: layout, Panes: make([]interface{}, 0)}
	cmd := Command{}
	cmd.Add("list-pane", "-t", window, "-F", "#{pane_current_path}|#{pane_current_command}|#{pane_active}")
	out, err := cmd.Execute(tmux, args)
	if err != nil {
		return nil, err
	}
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

	return &w, nil
}

// IsInsideTmux Check if we are inside tmux or not
func IsInsideTmux() bool {
	// Simply, if the TMUX is set in env, We are in it :)
	return os.Getenv("TMUX") != ""
}

// IsSessionExists check if a session name is available on tmux or not
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

// LoadSession loads session from a bytes slice
func LoadSession(data []byte) (*Session, error) {
	session := &Session{}
	if err := yaml.Unmarshal(data, session); err != nil {
		return nil, err
	}

	appRoot, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	u := user.New()

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
					p.Root = u.HomeDir + p.Root[1:]
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

// LoadSessionFromFile try to load a session from file
func LoadSessionFromFile(fileName string) (*Session, error) {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return LoadSession(content)
}

// SaveSessionToFile save a session struct to a file name
func SaveSessionToFile(sess []byte, fileName string) error {
	return ioutil.WriteFile(fileName, sess, 0644)
}

func init() {
	DefaultCmd = `echo "Default CMD"`
	IgnoredCmd = []string{os.Args[0]}
	if s := os.Getenv("SHELL"); s != "" {
		b := path.Base(s)
		IgnoredCmd = append(IgnoredCmd, s, b)
	}

	if v := os.Getenv("TMUX"); v != "" {
		tmuxENV = make([]string, len(os.Environ())-1)
		for _, e := range os.Environ() {
			if e != "TMUX="+v {
				tmuxENV = append(tmuxENV, e)
			}
		}
	} else {
		tmuxENV = os.Environ()
	}
}
