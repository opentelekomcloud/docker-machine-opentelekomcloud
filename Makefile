export GO111MODULE=on
export PATH:=/usr/local/go/bin:$(PATH)

default: test build
test: vet

fmt:
	@echo Running go fmt
	@go fmt

lint:
	@echo Running go lint
	@golint ./...

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

build:
	@echo Build driver
	@go build -o bin/docker-machine-driver-opentelecomcloud