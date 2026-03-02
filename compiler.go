package noirgo

import (
	"context"
	"github.com/YaniXIV/noir-go/internal/compiler"
)

// Compile compiles a Noir project at projectPath using a background context.
func Compile(projectPath string) (*compiler.Compilation, error) {
	return CompileWithContext(context.Background(), projectPath)
}

// CompileWithContext compiles a Noir project at projectPath with the provided context.
func CompileWithContext(ctx context.Context, projectPath string) (*compiler.Compilation, error) {
	e, err := New()
	if err != nil {
		return nil, err
	}
	defer e.CloseWithContext(ctx) // or defer e.Close(ctx) if you prefer

	return e.CompileWithContext(ctx, projectPath)
}

// Compile compiles a Noir project at projectPath using a background context.
func (e *Engine) Compile(projectPath string) (*compiler.Compilation, error) {
	return e.CompileWithContext(context.Background(), projectPath)
}

// CompileWithContext compiles a Noir project at projectPath with the provided context.
func (e *Engine) CompileWithContext(ctx context.Context, projectPath string) (*compiler.Compilation, error) {
	return compiler.CompileProgram(ctx, e.wm, projectPath)
}
