CFLAGS ?= -g -std=gnu11 -Wall -Wextra -Werror -Wno-shadow -fsanitize=address -fsanitize=undefined -fstack-protector-all -fno-omit-frame-pointer
LDLIBS ?= -lm

CLANG       = clang
GCC_NATIVE  = gcc-15
GCC_DOCKER  = docker run --rm -v "$(shell pwd)":/src -w /src gcc:15.2.0

compiler ?= $(СС)
RUN_CMD = ./build/main

ifeq ($(compiler), clang)
    CC = $(CLANG)
else ifeq ($(compiler), gcc)
    CC = $(GCC_NATIVE)
	CFLAGS += -fanalyzer -D_FORTIFY_SOURCE=2
else ifeq ($(compiler), docker)
    CC = $(GCC_DOCKER) gcc
	CFLAGS += -fanalyzer -D_FORTIFY_SOURCE=2
    RUN_CMD = $(GCC_DOCKER) ./build/main
endif

inspect:
	go run ./cmd/inspect -- $(path)

test:
	@go test ./so/...
	@go test ./internal/...

dist:
	@rm -rf dist
	@mkdir -p dist/solod/bin
	@go build -o dist/solod/bin/so ./cmd/so
	@tar -czf dist/solod.tar.gz -C dist solod
	@echo "Created dist/solod.tar.gz"

run-cases:
	@make run-cases-by pattern="testdata/lang/*/ testdata/std/*/"

run-cases-windows:
	@make run-cases-by CFLAGS="-g -std=gnu11 -Wall -Wextra -Werror -Wno-shadow -lm" pattern="testdata/lang/*/"

run-cases-by:
	@failed=0; \
	for dir in $(pattern); do \
		name=$${dir#testdata/}; \
		name=$${name%/}; \
		if make run-case name=$$name > /tmp/so_test_out.txt 2>&1; then \
			echo "PASS $$name"; \
		else \
			echo "FAIL $$name"; \
			cat /tmp/so_test_out.txt; \
			failed=1; \
		fi; \
	done; \
	rm -f /tmp/so_test_out.txt; \
	if [ $$failed -eq 0 ]; then \
		echo "PASS"; \
	else \
		echo "FAIL"; \
		exit 1; \
	fi

run-case:
	@rm -rf generated/$(name)
	@mkdir -p generated/$(name)
	@cp testdata/$(name)/dst/*.ext.[ch] generated/$(name)/ 2>/dev/null || true
	@go run ./cmd/so translate -o generated/$(name) testdata/$(name)/src
	@make run-c path=generated/$(name)

run-example:
	@mkdir -p example/$(name)/generated
	@rm -rf example/$(name)/generated/*
	@go run ./cmd/so translate -o example/$(name)/generated example/$(name)
	@rm -rf example/$(name)/generated/so

run-c:
	@mkdir -p build
	@$(CC) -O1 $(CFLAGS) -I$(path) -o build/main $(shell find $(path) -name "*.c") $(LDLIBS)
	@$(RUN_CMD)
	@rm -f build/main
