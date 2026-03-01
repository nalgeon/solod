CFLAGS = -g -std=gnu11 -Wall -Wextra -Werror -Wshadow -fsanitize=address -fsanitize=undefined -fstack-protector-all

example:
	@rm -rf generated/$(name)
	@mkdir -p generated/$(name)
	@go run ./cmd/so translate -o generated/$(name) tests/$(name)/src

inspect:
	go run ./cmd/inspect -- $(path)

runc:
	@mkdir -p build
	@gcc $(CFLAGS) -I$(path) -o build/main $(shell find $(path) -name "*.c")
	@./build/main
	@rm -f build/main

test:
	@go test ./internal/...

dist:
	@rm -rf dist
	@mkdir -p dist/solod/bin
	@go build -o dist/solod/bin/so ./cmd/so
	@tar -czf dist/solod.tar.gz -C dist solod
	@echo "Created dist/solod.tar.gz"