default:
	$(error please pick a target)

test:
	go tool staticcheck ./...
	go test -v ./...

release-patch: clean-git
	git tag $(shell go tool svu patch)
	git push
	git push --tags

release-minor: clean-git
	git tag $(shell go tool svu minor)
	git push
	git push --tags

ci: test

snapshot:
	rm -rf dist
	goreleaser release --snapshot --clean

clean-git:
	git diff --quiet

run:
	go run ./cmd/droopy/

build:
	go build ./cmd/droopy
