package compiler

func Compile(Input any) ([]byte, error) {
	err, v := runWasmCompiler()
	if err != nil {
		panic(err)
	}

	return v

}
