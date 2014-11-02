package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"
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
	Commands []string `yaml:"commands"`
	Focus    bool     `yaml:"focus,omitempty"`
	Root     string   `yaml:"root,omitempty"`
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
					panic(err) // TODO : Don't Panic!!
				}
				p := &Pane{}
				if err := yaml.Unmarshal(y, p); err != nil {
					panic(err) // TODO : Don't Panic!!
				}
				if p.Root == "" {
					p.Root = appRoot // Make sure each pane has a root
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

func GetHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
