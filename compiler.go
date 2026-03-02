package noirgo

import (
	"context"
	"noir-go/internal/compiler"
)

func Compile(projectPath string) (*compiler.Compilation, error) {
	return CompileWithContext(context.Background(), projectPath)
}

func CompileWithContext(ctx context.Context, projectPath string) (*compiler.Compilation, error) {
	e, err := New()
	if err != nil {
		return nil, err
	}
	defer e.CloseWithContext(ctx) // or defer e.Close(ctx) if you prefer

	return e.CompileWithContext(ctx, projectPath)
}

func (e *Engine) Compile(projectPath string) (*compiler.Compilation, error) {
	return e.CompileWithContext(context.Background(), projectPath)
}

func (e *Engine) CompileWithContext(ctx context.Context, projectPath string) (*compiler.Compilation, error) {
	return compiler.CompileProgram(ctx, e.wm, projectPath)
}
