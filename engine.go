package noirgo

import (
	"context"
	"github.com/YaniXIV/noir-go/internal/wasm"
)

// Engine owns the WASM runtime used to compile Noir projects.
type Engine struct {
	wm *wasm.WasmManager
}

// New constructs an Engine with a fresh WASM runtime.
func New() (*Engine, error) { // create engine (alloc runtime)
	wm, err := wasm.NewWasmManager()
	if err != nil {
		return nil, err
	}
	return &Engine{wm: wm}, nil
}

// NewEngine wraps an existing WasmManager (tests or custom configuration).
func NewEngine(wm *wasm.WasmManager) *Engine {
	return &Engine{wm: wm}
}

// Close releases the WASM runtime.
func (e *Engine) Close() error {
	return e.CloseWithContext(context.Background())
}

// CloseWithContext releases the WASM runtime with a context for cancellation.
func (e *Engine) CloseWithContext(ctx context.Context) error {
	return e.wm.Close(ctx)
}
