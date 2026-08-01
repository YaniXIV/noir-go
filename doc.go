// Package noirgo provides a Go-first interface for compiling Noir projects.
//
// The compiler runs via a bundled WASM bridge so consumers only need Go code
// and a path to a Noir project containing a Nargo.toml.
//
// Basic usage:
//
//	comp, err := noirgo.Compile("/path/to/noir-project")
//	if err != nil {
//		log.Fatal(err)
//	}
//	acir := comp.ACIR
//
// To execute a compiled program:
//
//	witness, err := noirgo.Execute(comp, inputs)
//	if err != nil {
//		log.Fatal(err)
//	}
//	_ = witness
package noirgo
