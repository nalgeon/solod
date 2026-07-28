CFLAGS_CORE = -O1 -g -std=gnu11 -Wall -Wextra -Werror -Wno-shadow -Wno-unused-label
CFLAGS ?= $(CFLAGS_CORE) -fsanitize=address -fsanitize=undefined -fstack-protector-all -fno-omit-frame-pointer
LDLIBS ?= -lm

CLANG = clang
GCC_NATIVE = gcc-15
GCC_DOCKER = docker run --rm -v "$(shell pwd)":/src -w /src gcc:15.2.0
RISCV64 = docker run --rm --platform linux/riscv64 -v "$(shell pwd)":/src -w /src solod/riscv64
I386 = docker run --rm --platform linux/i386 -v "$(shell pwd)":/src -w /src solod/i386
EMCC = emcc
ZIG = zig cc

mode =
OUT_NAME = main
RUN_CMD = ./build/main

# Set CC and CFLAGS based on the selected mode.
ifeq ($(mode), clang)
    CC = $(CLANG)
else ifeq ($(mode), gcc)
    CC = $(GCC_NATIVE)
	CFLAGS += -fanalyzer -D_FORTIFY_SOURCE=2
else ifeq ($(mode), docker)
    CC = $(GCC_DOCKER) gcc
	CFLAGS += -fanalyzer -D_FORTIFY_SOURCE=2
    RUN_CMD = $(GCC_DOCKER) ./build/main
else ifeq ($(mode), fast)
	CFLAGS = $(CFLAGS_CORE)
else ifeq ($(mode), bare)
	CC = $(ZIG)
	CFLAGS = $(CFLAGS_CORE) --target=wasm32-freestanding -nostdlib -Wl,--no-entry -Wl,--export=main -DSO_HEAP_SIZE=65536
	LDLIBS =
	OUT_NAME = main.wasm
	RUN_CMD = wasmtime --invoke main ./build/main.wasm 0 0
else ifeq ($(mode), riscv64)
	CC = $(RISCV64) gcc
	CFLAGS = $(CFLAGS_CORE)
	RUN_CMD = $(RISCV64) ./build/main
else ifeq ($(mode), i386)
	CC = $(I386) gcc
	CFLAGS = $(CFLAGS_CORE)
	RUN_CMD = $(I386) ./build/main
else ifeq ($(mode), wasm)
	CC = $(EMCC)
	CFLAGS = $(CFLAGS_CORE) -sSTANDALONE_WASM
	OUT_NAME = main.wasm
	RUN_CMD = wasmtime ./build/main.wasm
endif

# Preload mimalloc if available.
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    MIMALLOC_LIB := $(shell ls /opt/homebrew/lib/libmimalloc.dylib /usr/local/lib/libmimalloc.dylib 2>/dev/null | head -1)
    ifneq ($(MIMALLOC_LIB),)
        MIMALLOC_PRELOAD := DYLD_INSERT_LIBRARIES=$(MIMALLOC_LIB)
    endif
else ifeq ($(UNAME_S),Linux)
    MIMALLOC_LIB := $(shell ls /usr/lib/libmimalloc.so /usr/local/lib/libmimalloc.so 2>/dev/null | head -1)
    ifneq ($(MIMALLOC_LIB),)
        MIMALLOC_PRELOAD := LD_PRELOAD=$(MIMALLOC_LIB)
    endif
endif

inspect:
	go run ./cmd/inspect -- $(path)

test:
	@mkdir -p generated
	@go test ./so/...
	@go test ./internal/...

prepare-riscv64:
	@printf 'FROM alpine:edge\nRUN apk add --no-cache gcc musl-dev\n' \
		| docker build --platform=linux/riscv64 -t solod/riscv64 -
	@docker run --rm -it --platform=linux/riscv64 -v $(shell pwd):/src solod/riscv64 uname -m

prepare-i386:
	@printf 'FROM alpine:edge\nRUN apk add --no-cache gcc musl-dev\n' \
		| docker build --platform=linux/i386 -t solod/i386 -
	@docker run --rm -it --platform=linux/i386 -v $(shell pwd):/src solod/i386 uname -m

update-dst:
	make run-case name=$(name)
	cp generated/$(name)/main.* testdata/$(name)/dst
	go test -run TestTranslate/$(name) ./internal/compiler

# Overwrites the expected error of a testdata/bad case with the actual one.
# Updates every case if name is empty.
update-err:
	go test -run 'TestTranslateBad/$(name)' ./internal/compiler -update
	go test -run 'TestTranslateBad/$(name)' ./internal/compiler

# Runs tests in every testdata/* subdirectory, except testdata/bad
# (those cases must fail to translate, so there is nothing to run).
test-lang:
	@mkdir -p generated
	@failed=0; \
	for dir in $$(ls -d testdata/*/); do \
		name=$${dir#testdata/}; \
		name=$${name%/}; \
		if [ "$$name" = "bad" ]; then continue; fi; \
		if make run-case name=$$name > generated/so_test_out.txt 2>&1; then \
			echo "PASS $$name"; \
		else \
			echo "FAIL $$name"; \
			cat generated/so_test_out.txt; \
			failed=1; \
		fi; \
	done; \
	rm -f generated/so_test_out.txt; \
	if [ $$failed -eq 0 ]; then \
		echo "PASS"; \
	else \
		echo "FAIL"; \
		exit 1; \
	fi

# Runs tests in every stdlib package's "test" subdirectory.
test-std:
	@mkdir -p generated
	@failed=0; \
	for dir in $$(find so -type d -name test | sort); do \
		name=$${dir%/test}; \
		if make run-test name=$$name > generated/so_test_out.txt 2>&1; then \
			echo "PASS $$name"; \
		else \
			echo "FAIL $$name"; \
			cat generated/so_test_out.txt; \
			failed=1; \
		fi; \
	done; \
	rm -f generated/so_test_out.txt; \
	if [ $$failed -eq 0 ]; then \
		echo "PASS"; \
	else \
		echo "FAIL"; \
		exit 1; \
	fi

# Transpiles, compiles and runs a single test case in testdata/$(name),
# leaving the generated C in generated/$(name) for inspection.
run-case:
	@rm -rf generated/$(name)
	@mkdir -p generated/$(name)
	@cp testdata/$(name)/dst/*.ext.[ch] generated/$(name)/ 2>/dev/null || true
	@go run ./cmd/so translate -o generated/$(name) testdata/$(name)/src
	@make run-c path=generated/$(name)

# Transpiles, compiles and runs the tests in a package's "test" subdirectory
# (e.g. name=so/sync runs so/sync/test), leaving the generated C in
# generated/$(name)/test for inspection. It relies on the committed test
# runner (test/main.go); regenerate that with `so test` when tests change.
run-test:
	@rm -rf generated/$(name)/test
	@mkdir -p generated/$(name)/test
	@go run ./cmd/so translate -o generated/$(name)/test $(name)/test
	@make run-c path=generated/$(name)/test

run-c:
	@mkdir -p build
	@$(CC) $(CFLAGS) \
		-I$(path) \
		-o build/$(OUT_NAME) \
		$(shell find $(path) -name "*.c") \
		$(LDLIBS)
	@$(RUN_CMD)
	@rm -f build/$(OUT_NAME)

.PHONY: bench
bench:
	@cd $(name)/bench && go test -bench=. -benchmem
	@CFLAGS="-Ofast -march=native -flto -funroll-loops -DNDEBUG" \
	$(MIMALLOC_PRELOAD) \
	go run ./cmd/so bench -assert=off ./$(name)
