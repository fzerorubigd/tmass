package main

import (
	"errors"
	"io/ioutil"
	"os"
	"os/user"

	"gopkg.in/yaml.v2"
)

func LoadSession(data []byte) (*Session, error) {
	session := &Session{}
	if err := yaml.Unmarshal(data, session); err != nil {
		return nil, err
	}

	appRoot, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	home, err := GetHomeDir()
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

func GetHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}
