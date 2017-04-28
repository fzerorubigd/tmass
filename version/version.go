package version

import (
	"os"
	"runtime"
	"strconv"
	"text/template"
	"time"
)

var (
	Version   string
	GitCommit string
	GitState  string
	BuildDate string
)

var versionTemplate = `Version:     {{.Version}}
Git commit:  {{.GitCommit}}{{if eq .GitState "dirty"}}
Git State:   {{.GitState}}{{end}}
Built:       {{.BuildDate}}
Go version:  {{.GoVersion}}
OS/Arch:     {{.Os}}/{{.Arch}}
`

const CustomDateTime = "2006-01-02 15:04:05 -0700 MST"

type VersionInfo struct {
	Version   string
	GoVersion string
	GitCommit string
	GitState  string
	BuildDate string
	Os        string
	Arch      string
}

func New() *VersionInfo {
	var tu string

	if len(BuildDate) == 0 {
		tu = time.Now().Format(CustomDateTime)
	} else {
		i, err := strconv.ParseInt(BuildDate, 10, 64)
		if err != nil {
			panic(err)
		}

		tu = time.Unix(i, 0).String()
	}

	return &VersionInfo{
		Version:   Version,
		GoVersion: runtime.Version(),
		GitCommit: GitCommit,
		GitState:  GitState,
		BuildDate: tu,
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

func (i *VersionInfo) ShowVersion() {
	tmpl, err := template.New("version").Parse(versionTemplate)
	if err != nil {
		panic(err)
	}

	if err := tmpl.Execute(os.Stdout, i); err != nil {
		panic(err)
	}
}
