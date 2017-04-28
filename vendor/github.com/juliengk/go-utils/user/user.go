package user

import (
	"strconv"
	"syscall"
)

type User struct {
	Uid      string
	Gid      string
	Username string
	Name     string
	HomeDir  string
}

func New() *User {
	u := User{}

	u.Uid = strconv.Itoa(syscall.Getuid())
	u.Gid = strconv.Itoa(syscall.Getgid())

	username, exists := syscall.Getenv("USER")
	if exists {
		u.Username = username
	}

	homedir, exists := syscall.Getenv("HOME")
	if exists {
		u.HomeDir = homedir
	}

	return &u
}

func (u *User) IsRoot() bool {
	if u.Uid == "0" {
		return true
	}

	return false
}
