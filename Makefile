CFLAGS_CORE = -O1 -g -std=gnu11 -Wall -Wextra -Werror -Wno-shadow -Wno-unused-label
CFLAGS ?= $(CFLAGS_CORE) -fsanitize=address -fsanitize=undefined -fno-sanitize-recover=all -fstack-protector-all -fno-omit-frame-pointer
LDLIBS ?= -lm

CLANG = clang
GCC_NATIVE = gcc-15
GCC_DOCKER = docker run --rm -v "$(shell pwd)":/src -w /src gcc:15.2.0
RISCV64 = docker run --rm --platform linux/riscv64 -v "$(shell pwd)":/src -w /src solod/riscv64
I386 = docker run --rm --platform linux/i386 -v "$(shell pwd)":/src -w /src solod/i386
EMCC = emcc
ZIG = zig cc

mode =
OUT_EXT =
RUN_PREFIX =
RUN_SUFFIX =

# Set CC and CFLAGS based on the selected mode.
ifeq ($(mode), clang)
    CC = $(CLANG)
else ifeq ($(mode), gcc)
    CC = $(GCC_DOCKER) gcc
    RUN_PREFIX = $(GCC_DOCKER)
# The analyzer build drops the sanitizers on purpose. -fsanitize=nonnull-attribute
# is the runtime twin of -Wanalyzer-null-argument, and GCC does not report
# statically what it checks at run time.
else ifeq ($(mode), analyze)
    CC = $(GCC_DOCKER) gcc
	CFLAGS = $(CFLAGS_CORE) -fanalyzer -D_FORTIFY_SOURCE=2
    RUN_PREFIX = $(GCC_DOCKER)
else ifeq ($(mode), fast)
	CFLAGS = $(CFLAGS_CORE)
else ifeq ($(mode), bare)
	CC = $(ZIG)
	CFLAGS = $(CFLAGS_CORE) --target=wasm32-freestanding -nostdlib -Wl,--no-gc-sections -Wl,--no-entry -Wl,--export=main -DSO_HEAP_SIZE=65536
	LDLIBS =
	OUT_EXT = .wasm
	RUN_PREFIX = wasmtime --invoke main
	RUN_SUFFIX = 0 0
else ifeq ($(mode), riscv64)
	CC = $(RISCV64) gcc
	CFLAGS = $(CFLAGS_CORE)
	RUN_PREFIX = $(RISCV64)
else ifeq ($(mode), i386)
	CC = $(I386) gcc
	CFLAGS = $(CFLAGS_CORE)
	RUN_PREFIX = $(I386)
else ifeq ($(mode), wasm)
	CC = $(EMCC)
	CFLAGS = $(CFLAGS_CORE) -sSTANDALONE_WASM
	OUT_EXT = .wasm
	RUN_PREFIX = wasmtime
endif

# The name of the compiled binary. test-lang gives each case a name of its own,
# because the cases build in parallel.
bin = main$(OUT_EXT)
RUN_CMD = $(RUN_PREFIX) ./build/$(bin) $(RUN_SUFFIX)

# The transpiler command. test-lang overrides it with a prebuilt binary,
# because `go run` links the transpiler on every call.
SO = go run ./cmd/so

# The number of cases that test-lang runs in parallel.
jobs = $(shell getconf _NPROCESSORS_ONLN 2>/dev/null || echo 8)

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

# Runs tests in every testdata/* subdirectory, except two:
# - testdata/bad, because those cases must fail to translate.
# - testdata/freestanding, because it imports the stdlib.
#   (run with `make run-case name=freestanding mode=bare` instead)
# The cases run $(jobs) at a time. Each case gets a binary and a log of its own,
# and all cases share the transpiler binary that this target builds one time.
test-lang:
	@rm -rf generated/lang
	@mkdir -p generated/lang build
	@go build -o build/so ./cmd/so
	@if ls -d testdata/*/ | sed 's|testdata/||; s|/$$||' | grep -Evx 'bad|freestanding' | \
		xargs -P $(jobs) -I% make test-lang-case name=% SO=build/so; \
	then \
		echo "PASS"; \
	else \
		echo "FAIL"; \
		exit 1; \
	fi

# Runs one test-lang case. Prints the result and keeps the output in a log.
# The log makes the output of a failed case stay together, because the cases
# run in parallel.
test-lang-case:
	@if make run-case name=$(name) bin=$(name)$(OUT_EXT) > generated/lang/$(name).txt 2>&1; then \
		echo "PASS $(name)"; \
	else \
		echo "FAIL $(name)"; \
		sed 's|^|$(name): |' generated/lang/$(name).txt; \
		exit 1; \
	fi

# Runs the tests of every stdlib package with a "test" subdirectory. `so test`
# joins them into one program, so this is one translate, one compile and one
# run. The generated C stays in generated/std for inspection.
test-std:
	@rm -rf generated/std
	@mkdir -p generated/std
	@$(SO) translate -test -o generated/std ./so/...
	@make run-c path=generated/std

# Transpiles, compiles and runs a single test case in testdata/$(name),
# leaving the generated C in generated/$(name) for inspection.
run-case:
	@rm -rf generated/$(name)
	@mkdir -p generated/$(name)
	@cp testdata/$(name)/dst/*.ext.[ch] generated/$(name)/ 2>/dev/null || true
	@$(SO) translate -o generated/$(name) testdata/$(name)/src
	@make run-c path=generated/$(name)

# Transpiles, compiles and runs the tests in a package's "test" subdirectory
# (e.g. name=so/sync runs so/sync/test), leaving the generated C in
# generated/$(name)/test for inspection.
run-test:
	@rm -rf generated/$(name)/test
	@mkdir -p generated/$(name)/test
	@$(SO) translate -test -o generated/$(name)/test ./$(name)
	@make run-c path=generated/$(name)/test

run-c:
	@mkdir -p build
	@$(CC) $(CFLAGS) \
		-I$(path) \
		-o build/$(bin) \
		$(shell find $(path) -name "*.c") \
		$(LDLIBS)
	@$(RUN_CMD)
	@rm -f build/$(bin)

.PHONY: bench
bench:
	@cd $(name)/bench && go test -bench=. -benchmem
	@CFLAGS="-Ofast -march=native -flto -funroll-loops -DNDEBUG" \
	$(MIMALLOC_PRELOAD) \
	go run ./cmd/so bench -assert=off ./$(name)
