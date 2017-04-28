TARGETS := $(shell ls scripts | grep -vE 'clean|dev|help|release')

.dapper:
	@echo Downloading dapper
	@curl -sL https://releases.rancher.com/dapper/latest/dapper-`uname -s`-`uname -m|sed 's/v7l//'` > .dapper.tmp
	@@chmod +x .dapper.tmp
	@./.dapper.tmp -v
	@mv .dapper.tmp .dapper

.github-release:
	@echo Downloading github-release
	@curl -sL https://github.com/aktau/github-release/releases/download/v0.6.2/linux-amd64-github-release.tar.bz2 | tar xjO > .github-release.tmp
	@@chmod +x .github-release.tmp
	@./.github-release.tmp -v
	@mv .github-release.tmp .github-release

$(TARGETS): .dapper
	./.dapper $@

clean:
	@./scripts/clean

dev: .dapper
	./.dapper -m bind -s

help:
	@./scripts/help

release: .github-release
	./scripts/release

.DEFAULT_GOAL := ci

.PHONY: .dapper .github-release $(TARGETS) clean dev help release
