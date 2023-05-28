GO ?= go
GOLINT ?= golint
GOSEC ?= gosec
VERSION ?=$(shell git describe --tags --always)
PACKAGES = $(shell go list -f {{.Dir}} ./... | grep -v /vendor/)
DATE = $(shell date -R)
HOSTNAME=virtomize.com
NAMESPACE=uii
NAME=virtomize
BINARY=terraform-provider-${NAME}
OS_ARCH=linux_amd64
OS_ARCH_WIN=windows_amd64
TARGET_DIR_WIN := C:\Tools\Terraform\Plugins\${HOSTNAME}\${NAMESPACE}\${NAME}\${VERSION}\${OS_ARCH_WIN}

.PHONY: help
help: ## Show this help.
	@echo "Targets:"
	@grep -E '^[a-zA-Z\/_-]*:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\t%-20s: %s\n", $$1, $$2}'


.PHONY: build
build: ## build os executable.
ifeq ($(OS),Windows_NT)
	go build -o ${BINARY}.exe
else
	go build -o ${BINARY}
endif

.PHONY: src-fmt
src-fmt: ## format source code.
	gofmt -s -w ${PACKAGES}

.PHONY: release
release: ## release terraform build.
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

.PHONY: doc
doc: ## release terraform build.
	go generate ./...

.PHONY: install
install: build ## build terraform module.
ifeq ($(OS),Windows_NT)
	mkdir $(TARGET_DIR_WIN)
	copy $(BINARY).exe $(TARGET_DIR_WIN)
else
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
endif

.PHONY: test
test: ## Run tests.
	go test -race ./... -v -cover --coverprofile=coverage.out

.PHONY: cover
cover: ## Show test coverage.
	$(GO) tool cover -html=coverage.out

.PHONY: gosec
gosec: ## Run gosec static code security checker.
	$(GOSEC) ./...

.PHONY: testacc
testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

.PHONY: install-tools
install-tools: ## install dependencies if needed.
	$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin latest





