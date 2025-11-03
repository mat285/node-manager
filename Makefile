VERSION ?= v0.6.1
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
	gh release upload ${VERSION} build/gateway-server_linux_amd64 build/gateway-server_linux_arm64 build/gateway-server_darwin_arm64 gateway.service install.sh _config/example.yml --clobber

.PHONY: build
build:
	docker build \
	  --cache-from registry.k8s.nori.ninja/node-manager:latest \
	  --platform=linux/amd64 \
	  --tag registry.k8s.nori.ninja/node-manager:latest \
	  --tag registry.k8s.nori.ninja/node-manager:${GIT_SHA} \
	  --file Dockerfile \
	  --push \
	  .	

.PHONY: install
install:
	sh -c "$(curl -fsSL https://github.com/mat285/gateway/releases/download/${VERSION}/install.sh)"

.PHONY: install-local
install-local:
	./install.sh ${VERSION}

.PHONY: run
run:
	go run cmd/daemon/main.go --config-path=_config/example.yml