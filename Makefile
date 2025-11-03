VERSION ?= v0.8.1
GIT_SHA ?= $(shell git log --pretty=format:'%H' -n 1 2> /dev/null | cut -c1-8)

.PHONY: release-all
release-all: release
	VERSION=${VERSION} go run cmd/update/main.go ${VERSION}

.PHONY: release
release: build push-files

.PHONY: create-release
create-release:
	gh release create ${VERSION} --title "${VERSION}" --notes "Release ${VERSION} - Build ${GIT_SHA}" --generate-notes

.PHONY: push-files
push-files:
	@echo "Pushing files to GitHub release..."
	gh release upload ${VERSION} build/node-manager_linux_amd64 build/node-manager_linux_arm64 build/node-manager_darwin_arm64 node-manager.service install.sh _config/example.yml --clobber

.PHONY: build
build:
	./build.sh

.PHONY: install
install:
	sh -c "$(curl -fsSL https://github.com/mat285/node-manager/releases/download/${VERSION}/install.sh)"

.PHONY: install-local
install-local:
	./install.sh ${VERSION}

.PHONY: run
run:
	go run cmd/daemon/main.go --config-path=_config/example.yml