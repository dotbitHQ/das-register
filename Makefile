# build file
GOCMD=go
# Use -a flag to prevent code cache problems.
GOBUILD=$(GOCMD) build -ldflags -s -v -i

BIN_BINARY_NAME=das_register
register:
	$(GOBUILD) -o $(BIN_BINARY_NAME) cmd/main.go
	@echo "Build $(BIN_BINARY_NAME) successfully. You can run ./$(BIN_BINARY_NAME) now.If you can't see it soon,wait some seconds"

update:
	go mod tidy
	go mod vendor

# cli
CLI_BINARY_NAME=das_reward_cli
cli:
	$(GOBUILD) -o $(CLI_BINARY_NAME) cmd/cli/main.go