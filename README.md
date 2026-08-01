# noir-go

A Go library for compiling, executing, proving, and verifying [Noir](https://noir-lang.org/) circuits without requiring a local Noir toolchain.

The compiler and witness solver run via an embedded WASM binary (zst-compressed); proving and verification run as native Go via [gnark](https://github.com/Consensys/gnark). No Rust, no Nargo, no external toolchain, no CGo.

```sh
go get github.com/YaniXIV/noir-go
```

> **Status:** Early stage. Compile, execute, and Groth16 prove/verify all work end to end. The API is not stable yet and the library is not production ready. See [Limitations](#limitations) below -- most notably, circuits using black-box functions (range checks, hashes, signature verification, etc.) aren't provable yet. Contributions and feedback welcome.

---

## How it works

Most Go projects that need Noir have to shell out to the Nargo CLI or manage a local Noir toolchain. noir-go embeds the Noir compiler and ACVM solver as a WASM binary and runs them via [wazero](https://github.com/tetratelabs/wazero). Just `go get` and import.

```
Noir project (.nr + Nargo.toml)
        |
        v
  resolve + serialize      reads Nargo.toml, walks .nr sources, resolves
                            path/git dependencies, packs via msgpack
        |
        v
  embedded WASM compiler    the real Noir compiler (noirc_driver), cross
                            compiled to wasm32-wasip1, run via wazero
        |
        +--> Compile()  ->  ACIR + ABI + witness indices
        |
        +--> Execute()  ->  runs the same WASM module's ACVM solver against
        |                   concrete inputs -> a solved witness
        v
  gnark (pure Go)           decodes the ACIR, maps it onto a gnark R1CS,
                            and runs Groth16 -- Setup / Prove / Verify
```

Compile and Execute round-trip through the WASM boundary; Setup/Prove/Verify are pure Go and never touch WASM. Errors from the embedded compiler/solver -- a real Noir compile error, an unsatisfiable constraint -- surface as actual diagnostics, not generic decode failures.

---

## Usage

### Compile and execute

```go
package main

import (
	"fmt"
	"log"

	noirgo "github.com/YaniXIV/noir-go"
)

func main() {
	comp, err := noirgo.Compile("/path/to/noir-project")
	if err != nil {
		log.Fatalf("compile failed: %v", err)
	}

	fmt.Printf("ACIR size: %d bytes\n", len(comp.ACIR.Bytes))
	fmt.Printf("Public witnesses: %v\n", comp.PublicParamWitnesses)
	fmt.Printf("Private witnesses: %v\n", comp.PrivateParamWitnesses)

	inputs := make(map[uint32][32]byte)
	// ... populate one 32-byte, big-endian field element per witness index
	// in comp.PrivateParamWitnesses / comp.PublicParamWitnesses ...

	solved, err := noirgo.Execute(comp, inputs)
	if err != nil {
		log.Fatalf("execute failed: %v", err)
	}
	fmt.Printf("solved %d witness values\n", len(solved))
}
```

The project path should point to a Noir project directory containing a `Nargo.toml`.

### Setup, prove, verify

```go
pk, vk, err := noirgo.Setup(comp) // Groth16 trusted setup for this circuit (BN254)
if err != nil {
	log.Fatalf("setup failed: %v", err)
}

proof, err := noirgo.Prove(comp, pk, solved)
if err != nil {
	log.Fatalf("prove failed: %v", err)
}

ok, err := noirgo.Verify(vk, proof)
if err != nil {
	log.Fatalf("verify failed: %v", err)
}
fmt.Println("valid:", ok)
```

`pk`/`vk`/`proof` are gnark-native byte encodings -- persist them as-is (e.g. write `pk`/`vk` to disk once per circuit; a `proof` is produced per witness).

See [`examples/proveverify`](examples/proveverify) for a complete, runnable version of the full pipeline (`go run ./examples/proveverify`).

Package docs are available on [pkg.go.dev](https://pkg.go.dev/github.com/YaniXIV/noir-go).

---

## Limitations

- **Only `AssertZero` opcodes are provable.** Circuits using black-box functions -- range checks, `sha256`/`blake2s`/`keccak256`, ECDSA verification, Pedersen, etc. -- fail `Setup`/`Prove` with a clear "unsupported ACIR opcode" error rather than being silently mis-proved. A lot of real-world Noir programs use at least a range check, so this is the main practical gap today.
- **Groth16 over BN254 only.** No Plonk, no other curves, no Solidity/on-chain verifier export yet.
- **Library only**, no CLI.
- Public API is pre-1.0 and will still move.

---

## Roadmap

- [x] Compiler bridge (WASM via wazero)
- [x] Project resolution (Nargo.toml, path dependencies)
- [x] ACIR artifact + witness index metadata from compile results
- [x] Witness generation (ACVM solver via the embedded WASM module)
- [x] Prover integration (gnark Groth16/BN254, `AssertZero`-only circuits)
- [x] Verifier integration (gnark Groth16/BN254)
- [ ] Black-box function support (range checks, hashes, signatures, ...)
- [ ] Additional proving schemes / curves
- [ ] Stable public API

The goal is a Go-native ZK experience covering the full Noir lifecycle: compile, prove, and verify. With as few external toolchain dependencies as possible.
