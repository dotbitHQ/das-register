# build file
GO_BUILD=go build -ldflags -s -v

BIN_BINARY_NAME=das_register
register:
	$(GO_BUILD) -o $(BIN_BINARY_NAME) cmd/main.go
	@echo "Build $(BIN_BINARY_NAME) successfully. You can run ./$(BIN_BINARY_NAME) now.If you can't see it soon,wait some seconds"

update:
	go mod tidy

docker:
	docker build --network host -t dotbitteam/das-register:latest .

docker-publish:
	docker image push dotbitteam/das-register:latest

default:register
