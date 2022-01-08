GOPATH:=$(shell go env GOPATH)

.PHONY: format
## format: format files
format:
	@go install golang.org/x/tools/cmd/goimports@latest
	goimports -local github.com/taehoio -w .
	gofmt -s -w .
	go mod tidy

.PHONY: lint
## lint: check everything's okay
lint:
	@go install github.com/kyoh86/scopelint@latest
	golangci-lint run ./...
	scopelint --set-exit-status ./...
	go mod verify

.PHONY: help
## help: prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':'
