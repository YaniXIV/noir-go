# Paths (relative to server/)
RUSTLIB_COMPILER=noir-compile
TARGET_DIR=../internal/compiler
WASM_OUTPUT=$(TARGET_DIR)/noir-compile.wasm



build: build_compiler copy_wasm
cleanbuild: clean copy_wasm


build_compiler:
	cd $(RUSTLIB_COMPILER) && cargo build --release --target wasm32-wasip1


copy_wasm:
	cp $(RUSTLIB_COMPILER)/target/wasm32-wasip1/release/noir-compile.wasm $(WASM_OUTPUT)

clean: 
	cd $(RUSTLIB_COMPILER) && cargo clean && cargo build --release --target wasm32-wasip1
