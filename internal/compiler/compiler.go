package compiler

import (
	"bytes"
	"context"
	"encoding/binary"
	//"encoding/json"
	//"encoding/base64"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"log"
	"noir-go/internal/fs"
	"noir-go/internal/wasm"
	"unsafe"

	"github.com/tetratelabs/wazero"
)

type WireCompileResult struct {
	FormatVersion uint32 `msgpack:"format_version"`
	NoirVersion   string `msgpack:"noir_version"`
	AbiJSON       string `msgpack:"abi_json"`
	AcirString    string `msgpack:"acir_string"`
	AcirBytes     []int  `msgpack:"acir_bytes"`
	Hash          uint64 `msgpack:"hash"`
}

type AcirBlob []byte
type wasmBuf struct {
	ptr uint32
	len uint32
}

// simple compile function.
func Compile(projectPath string) {

	r := fs.NewResolver()

	r.Resolve(projectPath)
	projectData, err := r.Serialize()
	if err != nil {
		panic(err)
	}
	fmt.Println(projectData)

}

func CompileProgram(w *wasm.WasmManager, projectPath string) ([]byte, error) {
	obj, err := w.Get(wasm.Compiler)
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

	mod, errInstantiate := w.Instantiate(
		ctx, obj.Compiled, config,
	)
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
	//Serialization := mod.ExportedFunction("SerializationTest")
	fmt.Println("params:", Compiler.Definition().ParamTypes())
	fmt.Println("results:", Compiler.Definition().ResultTypes())

	if alloc == nil || Compiler == nil || dealloc == nil {
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

	/*
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
	*/

	//Call to Compiler. Args passed, size, length....
	//CompilerData, CompilerErr := Compiler.Call(ctx, uint64(ptr), uint64(size))
	outRes, err := alloc.Call(ctx, 8)
	if err != nil {
		return nil, err
	}
	outptr := outRes[0]

	_, CompilerErr := Compiler.Call(ctx, outptr, uint64(ptr), uint64(size))

	if CompilerErr != nil {
		fmt.Println("Compiler Failed", CompilerErr)
		return nil, nil
	}

	buf, ok := mem.Read(uint32(outptr), 8)
	if !ok {
		return nil, fmt.Errorf("failed to read WasmBuf")
	}

	retPtr := binary.LittleEndian.Uint32(buf[0:4])
	retLen := binary.LittleEndian.Uint32(buf[4:8])

	if retPtr == 0 || retLen == 0 {
		return nil, fmt.Errorf("compile failed (ptr=0,len=0)")
	}

	// 5) Read output bytes
	outBytes, ok := mem.Read(retPtr, retLen)
	log.Println(outBytes[0:5])
	copyBytes := make([]byte, len(outBytes))
	copy(copyBytes, outBytes)

	log.Println(outBytes[0:5], "first 5 outbytes")
	log.Println(copyBytes[0:5], "first 5 copybytes")

	//fmt.Println(outBytes, "<- output bytes")

	_, deallocErr := dealloc.Call(ctx, uint64(ptr), uint64(size))
	if deallocErr != nil {
		panic(deallocErr)
	}
	//CompilerOffset := uint32(CompilerData[0])
	//CompilerLength := uint32(CompilerData[1])

	//AcirBlob := make([]byte, CompilerLength)
	//AcirBlob, ok = mem.Read(CompilerOffset, CompilerLength)

	//fmt.Println(len(outBytes), "<- lenght of the outbytes")
	mod.Close(ctx)
	//printLogs(outputBuf)

	log.Println(outBytes[0:5], "first 5 outbytes")
	log.Println(copyBytes[0:5], "first 5 copybytes")

	var wire WireCompileResult
	err = msgpack.Unmarshal(copyBytes, &wire)
	if err != nil {
		return nil, err
	}

	/*
		var wire WireCompileResult
		err = json.Unmarshal(outBytes, &wire)
		if err != nil {
			fmt.Println(err)

		}
		fmt.Println(wire, "this is the wire")
		//fmt.Printf("Here are the Acir bytes From Golang!\n%v\n", AcirBlob)
	*/
	//fmt.Println(string(outBytes))
	//fmt.Println(wire, "this is the wire")
	fmt.Println("1", wire.AcirString)
	fmt.Println("2", wire.AcirBytes)
	//data, err := hex.DecodeString(wire.AcirBytes)
	//data, err := base64.StdEncoding.DecodeString(wire.AcirBytes)
	if err != nil {
		panic(err)
	}
	//fmt.Printf("2|%v|\n", string(data))
	/*
		if wire.AcirString == string(data) {
			fmt.Println("Yes hex and bin are the same.")
		} else {
			fmt.Println("Yes hex and bin are NOT the same.")

		}
	*/
	fmt.Println("3", wire.AbiJSON)
	//fmt.Println("2", wire.AcirB64)
	//fmt.Println("3", wire.FormatVersion)
	//fmt.Println("4", wire.Hash)
	//fmt.Println("5", wire.NoirVersion)
	return nil, nil
}
func printLogs(outBuf *bytes.Buffer) {
	fmt.Println("--- Start of wasm logs ---")
	fmt.Println(outBuf.String())
	fmt.Println("--- End of wasm logs ---")
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
