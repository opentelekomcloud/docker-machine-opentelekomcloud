export GO111MODULE=on
export PATH:=/usr/local/go/bin:$(PATH)
exec_path := /usr/local/bin/
exec_name := docker-machine-driver-otc

VERSION := 0.3.0b1


default: test build
test: vet acceptance

fmt:
	@echo Running go fmt
	@go fmt

lint:
	@echo Running go lint
	@golint --set_exit_status ./...

vet:
	@echo "go vet ."
	@go vet ./...

acceptance:
	@echo "Starting acceptance tests..."
	@go test ./driver/opentelekomcloud_test.go -race -covermode=atomic -coverprofile=coverage.txt -timeout 20m -v

acceptance-services:
	@echo "Starting acceptance tests for services..."
	@go test -v -race -timeout 60m ./driver/services

build: build-linux

build-linux:
	@echo "Build driver for Linux"
	@go build --trimpath -o bin/$(exec_name)

build-windows:
	@echo "Build driver for Windows"
	@GOOS=windows go build --trimpath -o bin/$(exec_name).exe

build-all: build-linux build-windows

install:
	@cp ./bin/$(exec_name) $(exec_path)
	@echo "Driver installed to $(exec_path)$(exec_name)"

release: build-all
	@tar -czf "./bin/$(exec_name)-$(VERSION)-linux-amd64.tgz" "./bin/$(exec_name)"
	@zip -qmD ./bin/$(exec_name)-$(VERSION)-win-amd64.zip ./bin/$(exec_name).exe
	@echo "Release versions are built"
