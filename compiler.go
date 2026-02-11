package noir

import (
	"fmt"
	"noir-go/internal/compiler"
	"noir-go/internal/fs"
)

func MustCompile(projectPath string) {

}

func Compile(projectPath string) error {

	w, err := compiler.NewWasmManager()
	if err != nil == nil {

	} else if w.runtime == nil {
		return fmt.Errorf("Wasm runtii")
	}

	data, err := compiler.CompileProgram(projectPath)

	if err != nil {
		return err
	}

	return nil
}
