GO = $(shell which go)
GODOC = $(shell which godoc)
GOLINT = $(shell which golangci-lint)
GOTEST = $(shell which gotest)
TARGET_DIR = $(shell pwd)/target

ifeq ($(GOTEST),)
GOTEST = $(GO) test
endif

.PHONY: clean
clean:
	@rm -Rf $(TARGET_DIR)

.PHONY: configure
configure: install-hooks
	@$(GO) install github.com/rakyll/gotest@latest
	@$(GO) install golang.org/x/tools/cmd/godoc@latest

.PHONY: godoc
godoc:
	@echo Godoc: "http://localhost:6060/pkg/github.com/mgiaccone/tripswitch\n"
	@$(GODOC) -http=:6060

.PHONY: install-hooks
install-hooks:
	@ln -s $(PWD)/scripts/pre-commit .git/hooks/pre-commit

.PHONY: lint
lint:
	@$(GOLINT) run ./...

.PHONY: target
target:
	@mkdir -p $(TARGET_DIR)/reports

.PHONY: qa
qa: lint test

.PHONY: test
test: clean target
	@$(GOTEST) -coverprofile=$(TARGET_DIR)/reports/coverage.out \
		-count=1 \
		-failfast \
		-race \
		-v ./...
	@$(GO) tool cover -func=$(TARGET_DIR)/reports/coverage.out
	@$(GO) tool cover -html=$(TARGET_DIR)/reports/coverage.out -o $(TARGET_DIR)/reports/coverage.html
	@echo "\nCoverage report: file://$(PWD)/target/reports/coverage.html"
