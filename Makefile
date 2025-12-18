default:
	$(error please pick a target)

test:
	$(shell go env GOPATH)/bin/staticcheck ./...
	go test -v ./...

release-patch: clean-git
	git tag $(shell svu patch)
	git push
	git push --tags

release-minor: clean-git
	git tag $(shell svu minor)
	git push
	git push --tags

install-tools:
	go install github.com/caarlos0/svu@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

ci: install-tools test

snapshot:
	rm -rf dist
	goreleaser release --snapshot --clean

clean-git:
	git diff --quiet

run:
	go run ./cmd/droopy/

build:
	go build -ldflags="-X main.version=$(shell git tag | tail -1)" ./cmd/droopy
