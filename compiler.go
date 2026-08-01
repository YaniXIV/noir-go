package noirgo

import (
	"context"

	"github.com/YaniXIV/noir-go/internal/compiler"
	"github.com/YaniXIV/noir-go/result"
)

// Compile compiles a Noir project at projectPath using a background context.
func Compile(projectPath string) (*result.Compilation, error) {
	return CompileWithContext(context.Background(), projectPath)
}

// CompileWithContext compiles a Noir project at projectPath with the provided context.
func CompileWithContext(ctx context.Context, projectPath string) (*result.Compilation, error) {
	e, err := New()
	if err != nil {
		return nil, err
	}
	defer e.CloseWithContext(context.Background())

	return e.CompileWithContext(ctx, projectPath)
}

// Compile compiles a Noir project at projectPath using a background context.
func (e *Engine) Compile(projectPath string) (*result.Compilation, error) {
	return e.CompileWithContext(context.Background(), projectPath)
}

// CompileWithContext compiles a Noir project at projectPath with the provided context.
func (e *Engine) CompileWithContext(ctx context.Context, projectPath string) (*result.Compilation, error) {
	return compiler.CompileProgram(ctx, e.wm, projectPath)
}

// Execute runs a compiled Noir program using a background context.
func Execute(comp *result.Compilation, inputs map[uint32][32]byte) (map[uint32][32]byte, error) {
	return ExecuteWithContext(context.Background(), comp, inputs)
}

// ExecuteWithContext runs a compiled Noir program with the provided context.
func ExecuteWithContext(ctx context.Context, comp *result.Compilation, inputs map[uint32][32]byte) (map[uint32][32]byte, error) {
	e, err := New()
	if err != nil {
		return nil, err
	}
	defer e.CloseWithContext(context.Background())

	return e.ExecuteWithContext(ctx, comp, inputs)
}
