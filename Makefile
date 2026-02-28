CFLAGS = -g -std=gnu11 -Wall -Wextra -Werror -Wshadow -fsanitize=address -fsanitize=undefined -fstack-protector-all

example:
	@rm -f generated/$(name)/*
	@mkdir -p generated/$(name)
	@go run ./cmd/so translate tests/$(name)/src -o generated/$(name)

inspect:
	go run ./cmd/inspect -- $(path)

runc:
	@mkdir -p build
	@gcc $(CFLAGS) -Iinternal/compiler/builtin -o build/main $(path)
	@./build/main
	@rm -f build/main

test:
	@go test ./internal/...

dist:
	@rm -rf dist
	@mkdir -p dist/soan/bin
	@go build -o dist/soan/bin/so ./cmd/so
	@tar -czf dist/soan.tar.gz -C dist soan
	@echo "Created dist/soan.tar.gz"
