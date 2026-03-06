# noir-go

A Go library for compiling [Noir](https://noir-lang.org/) circuits without requiring a local Noir toolchain.

The compiler runs via an embedded WASM binary (zst-compressed), so there are no external dependencies. Just a standard Go module.

```sh
go get github.com/YaniXIV/noir-go
```

> **Status:** Early stage. The compiler bridge works. The API is not stable yet and the library is not production ready. Contributions and feedback welcome.

---

## How it works

Most Go projects that need Noir have to shell out to the Nargo CLI or manage a local Noir toolchain. noir-go embeds the Noir compiler as a WASM binary and runs it via [wazero](https://github.com/tetratelabs/wazero). No Rust, no Nargo, no external toolchain required. Just `go get` and import.

The Go side resolves the Noir project (reads `Nargo.toml`, finds `.nr` sources, resolves path dependencies), serializes it via msgpack, and calls into the WASM compiler. The result is decoded into Go-native types containing the ACIR artifact, ABI, and witness indices.

---

## Usage

```go
package main

import (
	"fmt"
	"log"

	noirgo "github.com/YaniXIV/noir-go"
)

func main() {
	projectPath := "/path/to/noir-project"

	comp, err := noirgo.Compile(projectPath)
	if err != nil {
		log.Fatalf("compile failed: %v", err)
	}

	fmt.Printf("ACIR size: %d bytes\n", len(comp.ACIR))
	fmt.Printf("Public witnesses: %v\n", comp.PublicParamWitnesses)
	fmt.Printf("Private witnesses: %v\n", comp.PrivateParamWitnesses)
}
```

The `projectPath` should point to a Noir project directory containing a `Nargo.toml`.

Package docs are available on [pkg.go.dev](https://pkg.go.dev/github.com/YaniXIV/noir-go).

---

## Roadmap

- [x] Compiler bridge (WASM via wazero)
- [x] Project resolution (Nargo.toml, path dependencies)
- [x] ACIR artifact + witness index metadata from compile results
- [ ] Witness generation
- [ ] Prover integration
- [ ] Verifier integration
- [ ] Stable public API

The goal is a Go-native ZK experience covering the full Noir lifecycle: compile, prove, and verify. With as few external toolchain dependencies as possible.
