.PHONY: clean gox tmass

all: tmass

tmass: clean gox
	gox -os="linux darwin windows" -arch="amd64" -output="./dist/tmass-{{.OS}}-{{.Arch}}"

gox:
	command -v gox 1> /dev/null || GO111MODULE=off go get github.com/mitchellh/gox

clean:
	rm -rf ./dist

