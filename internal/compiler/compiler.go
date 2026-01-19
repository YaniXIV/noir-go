package compiler

import (
	"bytes"
	"context"
	"fmt"
	"noir-go/internal/fs"
	"unsafe"

	"github.com/tetratelabs/wazero"
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

	//cast uint64 cause wazero api boundry
	size := uint64(len(projectData))

	results, err := alloc.Call(ctx, size)
	if err != nil {
		return err
	}

	//get pointer result cast back to uint32
	ptr := uintptr(results[0]) // wasm32 → u32 pointer
	writeBytes(ptr, projectData)
	CompilerData, CompilerErr := fn.Call(ctx, uint64(ptr), size)
	if CompilerErr != nil {
		return CompilerErr
	}
	CompilerPtr := CompilerData[0]

	fmt.Printf("%x <-- HERE IS THE POINTER TO MEM Golang Side ", ptr)
	fmt.Println(size, " <-- here is the size of the allocated data! Golang Side ")

	fmt.Println("--- Start of wasm logs ---")
	fmt.Println(outputBuf.String())
	fmt.Println("--- End of wasm logs ---")

	return nil
}

// read the data back out, need lenght need ptr.... what else? hmm.
func readBytes(addr uintptr, size int) ([]byte, error) {
	if size < 1 {
		return nil, fmt.Errorf("data size cannot be < 1")
	}
	var AcirBlob []byte = make([]byte, size)
	data := unsafe.Slice((*byte)(unsafe.Pointer(addr)), size)
	copy(data, AcirBlob)
	if len(AcirBlob) < 1 {
		return nil, fmt.Errorf("Error writting data")
	}
	return AcirBlob, nil

}

// probably should do some size verification here, seems dangerous but yolo I guess.
func writeBytes(addr uintptr, data []byte) {
	dst := unsafe.Slice((*byte)(unsafe.Pointer(addr)), len(data))
	copy(dst, data)
}
