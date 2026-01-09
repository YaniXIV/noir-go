package compiler

import (
	"noir-go/internal/fs"
)

// simple compile function.
func Compile(projectPath string) {
	runWasmCompiler(noirWasm)

	r := fs.NewResolver()

	r.Resolve(projectPath)
	projectData, err := r.Serialize()
	if err != nil {
		panic(err)
	}

	runWasmCompiler(projectData)

}
