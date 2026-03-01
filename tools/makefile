# Paths (relative to server/)
RUSTLIB_COMPILER := noir-compile
TARGET_DIR       := ../internal/wasm

WASM_NAME        := noir-compile.wasm
RUST_TARGET      := wasm32-wasip1
PROFILE          := release
RUST_OUT_DIR     := $(RUSTLIB_COMPILER)/target/$(RUST_TARGET)/$(PROFILE)

WASM_SRC         := $(RUST_OUT_DIR)/$(WASM_NAME)
WASM_OUT         := $(TARGET_DIR)/$(WASM_NAME)

ZSTD_LEVEL       := 19
WASM_ZST_SRC     := $(WASM_SRC).zst
WASM_ZST_OUT     := $(WASM_OUT).zst

.PHONY: default build rebuild build_raw rebuild_raw \
        build_compiler build_compiler_clean \
        compress copy_zst copy_wasm clean

# Default: fast incremental build + zstd artifact
default: build

# --- Main entrypoints ---

# Incremental build + compress + copy .zst
build: build_compiler compress copy_zst

# Full clean rebuild (your "animation") + compress + copy .zst
rebuild: build_compiler_clean compress copy_zst

# Incremental build + copy raw .wasm (no compression)
build_raw: build_compiler copy_wasm

# Full clean rebuild + copy raw .wasm (no compression)
rebuild_raw: build_compiler_clean copy_wasm

# --- Build steps ---

build_compiler:
	cd $(RUSTLIB_COMPILER) && cargo build --release --target $(RUST_TARGET)

build_compiler_clean:
	cd $(RUSTLIB_COMPILER) && cargo clean && cargo build --release --target $(RUST_TARGET)

# --- Artifact steps ---

# Compress in-place in Rust output dir (overwrites .zst each time)
compress:
	cd $(RUST_OUT_DIR) && zstd -$(ZSTD_LEVEL) -q -f $(WASM_NAME)

copy_zst:
	mkdir -p $(TARGET_DIR)
	cp $(WASM_ZST_SRC) $(WASM_ZST_OUT)

copy_wasm:
	mkdir -p $(TARGET_DIR)
	cp $(WASM_SRC) $(WASM_OUT)

# --- Cleanup ---

clean:
	cd $(RUSTLIB_COMPILER) && cargo clean
