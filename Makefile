export GO111MODULE=on
export PATH:=/usr/local/go/bin:$(PATH)
export bin_path=/usr/local/bin/
export exec_name=docker-machine-driver-opentelekomcloud

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
	@echo "Starting acceptance tests..."
	@go test ./... -race -covermode=atomic -coverprofile=coverage.txt

build: build-linux

build-linux:
	@echo "Build driver for Linux"
	@go build --trimpath -o bin/${exec_name}

build-windows:
	@echo "Build driver for Windows"
	@GOOS=windows go build --trimpath -o bin/${exec_name}.exe

build-all: build-linux build-windows

install:
	@cp ./bin/${exec_name} ${bin_path}
	@echo "Driver installed to ${bin_path}${exec_name}"
