# noir-go

A personal project to make Noir usable directly from Go. The goal is a full Go library/SDK that covers the Noir lifecycle: compile, prove, and verify.

This is early and not production ready. The compiler bridge works, but it still needs hardening. The API is not stable yet. Under the hood, the compiler runs via WASM to keep the toolchain portable while presenting a Go-first API.

If you are interested, contributions and feedback are welcome.

This is a free-time project.

Future direction: a Go-native ZK experience with as few external dependencies as possible, ideally as simple as `go get github.com/YaniXIV/noir-go` and `import "github.com/YaniXIV/noir-go"`.

# Compile Example

Below is a minimal example showing how to compile a Noir project from Go. The `projectPath` should point to a Noir project directory that contains a `Nargo.toml`.

# Compile Example

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

# File Tree

noir-go
├── compiler.go
├── engine.go
├── go.mod
├── go.sum
├── internal
│   ├── compiler
│   │   ├── bridge.go
│   │   ├── compiler_test.go
│   │   ├── compiler.go
│   │   ├── noirtest
│   │   ├── types.go
│   │   └── wire.go
│   ├── fs
│   │   ├── fs_test.go
│   │   ├── nargo.go
│   │   ├── project.go
│   │   └── resolver.go
│   └── wasm
│       ├── loaders.go
│       ├── noir-compile.wasm.zst
│       ├── wasm_test.go
│       └── wasm.go
├── LICENSE
├── noirgo_test.go
├── out
├── README.md
├── tools
│   ├── experimentation
│   │   ├── Nargo.toml
│   │   ├── src
│   │   ├── target
│   │   └── time.txt
│   ├── makefile
│   ├── noir-compile
│   │   ├── Cargo.lock
│   │   ├── Cargo.toml
│   │   ├── FAILURE_POINTS.md
│   │   ├── moreNotes.md
│   │   ├── note.md
│   │   ├── src
│   │   └── target
│   └── noir-prover
│       ├── Cargo.lock
│       ├── Cargo.toml
│       ├── err
│       ├── src
│       └── target
└── types
    ├── field.go
    └── types.go
