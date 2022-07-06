GO = $(shell which go)
GOTEST = $(shell which gotest)
TARGET_DIR = $(shell pwd)/target

ifeq ($(GOTEST),)
GOTEST = $(GO) test
endif

.PHONY: clean
clean:
	@rm -Rf $(TARGET_DIR)

.PHONY: target
target:
	@mkdir -p $(TARGET_DIR)/bin
	@mkdir -p $(TARGET_DIR)/test

.PHONY: test
test: target
	@$(GOTEST) -coverprofile=$(TARGET_DIR)/test/coverage.out \
		-count=1 \
		-failfast \
		-race \
		-v ./...
	@$(GO) tool cover -func=$(TARGET_DIR)/test/coverage.out
	@$(GO) tool cover -html=$(TARGET_DIR)/test/coverage.out -o $(TARGET_DIR)/test/coverage.html
