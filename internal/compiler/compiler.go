package compiler

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"noir-go/internal/fs"
	"unsafe"

	"github.com/tetratelabs/wazero"
)

type AcirBlob []byte

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

func (w *WasmManager) CompileProgram(projectPath string) ([]byte, error) {
	obj, err := w.Get(Compiler)
	if obj == nil {
		fmt.Println("OBJECT IS INVALID")
	}
	if err != nil {
		return nil, err
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

	//resolver
	r := fs.NewResolver()
	r.Resolve(projectPath)
	projectData, errSerialize := r.Serialize()
	if errSerialize != nil {
		panic(errSerialize)
	}
	//fmt.Println(projectData, " <-- Serialized Project data!!! ")

	//funciton exports
	alloc := mod.ExportedFunction("alloc")
	dealloc := mod.ExportedFunction("dealloc")
	Compiler := mod.ExportedFunction("compile_wasm")
	Serialization := mod.ExportedFunction("SerializationTest")

	if alloc == nil || Compiler == nil || dealloc == nil || Serialization == nil {
		return nil, fmt.Errorf("exported Function Error ")
	}

	// Call to alloc rust space.
	size := uint64(len(projectData))
	log.Println("length", size)
	results, err := alloc.Call(ctx, size)
	if err != nil {
		return nil, err
	}
	log.Println("Bytes Allocated!", size)
	ptr := results[0]
	mem := mod.Memory()
	ok := mem.Write(uint32(ptr), projectData)
	if !ok {
		return nil, fmt.Errorf("Write error to wasm mem.")
	}

	//Call to Serialization test. Args passed, size, length.
	SerializationData, serializationErr := Serialization.Call(ctx, uint64(ptr), uint64(size))

	log.Println("Serialization Passes?")
	if serializationErr != nil {
		fmt.Println("Serialization Error!")
		return nil, serializationErr

	}

	if SerializationData[0] != 0 {
		fmt.Println("function didn't reach the end! weird")
		fmt.Println(SerializationData[0])
	}
	fmt.Println("What is going on? ", SerializationData[0])

	/*
		//Call to Compiler. Args passed, size, length....
		CompilerData, CompilerErr := Compiler.Call(ctx, uint64(ptr), uint64(size))
		if CompilerErr != nil {
			return nil, CompilerErr
		}
		CompilerOffset := uint32(CompilerData[0])
		CompilerLength := uint32(CompilerData[1])

		AcirBlob := make([]byte, CompilerLength)
		AcirBlob, ok = mem.Read(CompilerOffset, CompilerLength)
		fmt.Printf("%x <-- HERE IS THE POINTER TO MEM Golang Side ", ptr)
		fmt.Println(size, " <-- here is the size of the allocated data! Golang Side ")
	*/

	mod.Close(ctx)
	fmt.Println("--- Start of wasm logs ---")
	fmt.Println(outputBuf.String())
	fmt.Println("--- End of wasm logs ---")

	//fmt.Printf("Here are the Acir bytes From Golang!\n%v\n", AcirBlob)

	return nil, nil
}

// Don't use this function
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

// Don't use this function
func writeBytes(addr uintptr, data []byte) {
	dst := unsafe.Slice((*byte)(unsafe.Pointer(addr)), len(data))
	copy(dst, data)
}
