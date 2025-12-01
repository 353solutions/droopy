default:
	$(error please pick a target)

test:
	$(shell go env GOPATH)/bin/staticcheck ./...
	go test -v ./...

release-patch:
	$(shell svu --strip-prefix patch > version.txt)
	git add version.txt
	git commit -m 'bump version' version.txt
	git tag v$(shell cat version.txt)
	git push
	git push --tags

release-minor:
	$(shell svu --strip-prefix minor > version.txt)
	git add version.txt
	git commit -m 'bump version' version.txt
	git tag v$(shell cat version.txt)
	git push
	git push --tags

install-tools:
	go install github.com/caarlos0/svu@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

ci: install-tools test
	
