export GO111MODULE=on
export PATH:=/usr/local/go/bin:$(PATH)
export bin_path=/usr/local/bin/

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
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

acceptance:
	@go test ./...

build: build-linux

build-linux:
	@echo "Build driver for Linux"
	@go build --trimpath -o bin/docker-machine-driver-opentelekomcloud

build-windows:
	@echo "Build driver for Windows"
	@GOOS=windows go build --trimpath -o bin/docker-machine-driver-opentelekomcloud.exe

build-all: build-linux build-windows

install:
	@cp bin/docker-machine-driver-opentelekomcloud ${bin_path}
	@echo "Driver installed to ${bin_path}"
