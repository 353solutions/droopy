default:
	$(error please pick a target)

test:
	go tool staticcheck ./...
	go test -v ./...


release-patch:
	$(shell go tool svu --strip-prefix patch > version.txt)
	git add version.txt
	git commit -m 'bump version' version.txt
	git tag v$(shell cat version.txt)
	git push
	git push --tags

release-minor:
	$(shell go tool svu --strip-prefix minor > version.txt)
	git add version.txt
	git commit -m 'bump version' version.txt
	git tag $(shell cat version.txt)
	git push
	git push --tags
