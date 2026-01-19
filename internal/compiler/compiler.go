package compiler

import (
	"bytes"
	"context"
	"fmt"
	"github.com/tetratelabs/wazero"
	"noir-go/internal/fs"
)

// simple compile function.
func Compile(projectPath string) {
	runWasmCompiler(noirCompilerWasm)

	r := fs.NewResolver()

	r.Resolve(projectPath)
	projectData, err := r.Serialize()
	if err != nil {
		panic(err)
	}

	runWasmCompiler(projectData)

}

func (w *WasmManager) CompileProgram(projectPath string) error {
	obj, err := w.Get(Compiler)
	if obj == nil {
		fmt.Println("OBJECT IS INVALID")
	}
	if err != nil {
		return err
	}
	ctx := context.Background()
	outputBuf := new(bytes.Buffer)
	config := wazero.NewModuleConfig().
		WithStdout(outputBuf).
		WithStderr(outputBuf)

	mod, errInstantiate := w.runtime.InstantiateModule(ctx, obj.Compiled, config)
	if errInstantiate != nil {
		fmt.Println("ERROR HERE LINE 40")
		panic(errInstantiate)
	}
	//defer mod.Close(ctx)
	r := fs.NewResolver()
	r.Resolve(projectPath)
	projectData, errSerialize := r.Serialize()
	if errSerialize != nil {
		panic(errSerialize)
	}
	fmt.Println(projectData, " <-- Serialized Project data!!! ")

	alloc := mod.ExportedFunction("alloc")
	fn := mod.ExportedFunction("test_compile_wasm_go")

	if alloc == nil || fn == nil {
		return fmt.Errorf("exported Function Error ")
	}

	size := uint64(len(projectData))

	results, err := alloc.Call(ctx, size)
	if err != nil {
		return err
	}

	ptr := uint32(results[0]) // wasm32 → u32 pointer
	fmt.Printf("%x <-- HERE IS THE POINTER TO MEM Golang Side ", ptr)
	fmt.Println(size, " <-- here is the size of the allocated data! Golang Side ")

	fmt.Println("--- Start of wasm logs ---")
	fmt.Println(outputBuf.String())
	fmt.Println("--- End of wasm logs ---")

	return nil
}
