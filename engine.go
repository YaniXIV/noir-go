package noirgo

import (
	"context"
	"noir-go/internal/wasm"
)

type Engine struct {
	wm *wasm.WasmManager
}

func New() (*Engine, error) { // create engine (alloc runtime)
	wm, err := wasm.NewWasmManager()
	if err != nil {
		return nil, err
	}
	return &Engine{wm: wm}, nil
}

// If you want to inject a prebuilt WasmManager (tests/custom config)
func NewEngine(wm *wasm.WasmManager) *Engine {
	return &Engine{wm: wm}
}
func (e *Engine) Close() error {
	return e.CloseWithContext(context.Background())
}
func (e *Engine) CloseWithContext(ctx context.Context) error {
	return e.wm.Close(ctx)
}
