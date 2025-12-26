use nargo::parse_all;
use noirc_driver::{CompileOptions, compile_main, file_manager_with_stdlib, prepare_crate};
use noirc_frontend::hir::Context;
use std::path::Path;

fn main() {
    println!("Hello, world!");

    compile_from_memory(
        "fn main(x: Field, y: pub Field) {
	assert(x != y);
}",
    )
}

#[unsafe(no_mangle)]

pub extern "C" fn test_compile_wasm_go() {
    compile_from_memory(
        "fn main(x: Field, y: pub Field) {
	assert(x != y);",
    )
}

pub fn compile_from_memory(source: &str) {
    // 1. Virtual filesystem + stdlib
    let mut fm = file_manager_with_stdlib(Path::new(""));
    fm.add_file_with_source(Path::new("/main.nr"), source.to_string())
        .unwrap();

    // 2. Parse
    let parsed_files = parse_all(&fm);

    // 3. Compiler context
    let mut context = Context::new(fm, parsed_files);

    // 4. Prepare crate
    let crate_id = prepare_crate(&mut context, Path::new("/main.nr"));

    // 5. Compile
    let options = CompileOptions::default();
    let result = compile_main(&mut context, crate_id, &options, None);

    println!("{:?}", result);
}
