PKG        := ./...
TEST_FLAGS := -cover
LINT_CMD   := golangci-lint

.PHONY: all test lint clean help

all: test lint

test:
	go test $(TEST_FLAGS) $(PKG)

lint:
	$(LINT_CMD) run ./...
