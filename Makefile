GO = $(shell which go)
GOTEST = $(shell which gotest)
GOLINT = $(shell which golangci-lint)
TARGET_DIR = $(shell pwd)/target

ifeq ($(GOTEST),)
GOTEST = $(GO) test
endif

.PHONY: clean
clean:
	@rm -Rf $(TARGET_DIR)

.PHONY: godoc
godoc:

.PHONY: lint
lint:
	@$(GOLINT) run ./...

.PHONY: target
target:
	@mkdir -p $(TARGET_DIR)/reports

.PHONY: test
test: clean target
	@$(GOTEST) -coverprofile=$(TARGET_DIR)/reports/coverage.out \
		-count=1 \
		-failfast \
		-race \
		-v ./...
	@$(GO) tool cover -func=$(TARGET_DIR)/reports/coverage.out
	@$(GO) tool cover -html=$(TARGET_DIR)/reports/coverage.out -o $(TARGET_DIR)/reports/coverage.html
