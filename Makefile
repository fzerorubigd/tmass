export GO=$(which go)
export ROOT=$(realpath $(dir $(lastword $(MAKEFILE_LIST))))
export BIN=$(ROOT)/bin

.PHONY: all tmass restore clean purge

tmass: $(BIN)/gb
	$(BIN)/gb build

all: clean restore tmass

restore: $(BIN)/gb
	$(BIN)/gb vendor restore 

gb:
	GOPATH=/tmp GOBIN=$(ROOT)/bin go get -v github.com/constabulary/gb/...
	rm -rf /tmp/src/github.com/constabulary/gb/

clean:
	rm -rf ./pkg ./vendor/pkg 
	rm -f $(BIN)/tmass

purge: clean 
	rm -rf ./vendor/src

$(BIN)/gb:
	[ -f $(BIN)/gb ] || make gb

update: $(BIN)/gb
	$(BIN)/gb vendor update --all
	
