.PHONY: build
build:
	go build -o git-org-manager .

.PHONY: run
run: build
	./git-org-manager

.PHONY: test
test:
	go test -v ./...
