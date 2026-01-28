use acir::circuit::Program;
use acvm::acir::circuit::ExpressionWidth;
use nargo::parse_all;
use noirc_driver::{CompileOptions, compile_main, file_manager_with_stdlib, prepare_crate};
use noirc_frontend::hir::Context;
use serde::Serialize;
use std::collections::HashMap;
use std::path::Path;
use std::ptr;

#[derive(Serialize)]
struct MyMap(HashMap<String, String>);

fn main() {
    //println!("Hello, world!");

    //compile_from_memory("fn main(x: Field, y: pub Field) {assert(x != y);}",)
    //compile_from_memory(
    //  "fn main() {
    //   let mut acc: u128 = 0;
    // let n: u128 = 1_000_000;

    //for i in 0..n {
    //  acc = acc + (i * i) % 1234567;
    // acc = acc ^ ((i + acc) % 987654321);
    //}

    // assert(acc == acc);
    //}",
    //  )

    let mut map: HashMap<String, String> = HashMap::new();
    map.insert(
        "/main.nr".to_string(),
        "fn main() {
    let mut acc: u128 = 0;
    let n: u128 = 1_000_000;

    for i in 0..n {
        acc = acc + (i * i) % 1234567;
        acc = acc ^ ((i + acc) % 987654321);
    }

    assert(acc == acc);
}"
        .to_string(),
    );
    let my_map = MyMap(map);

    let mut bytes = rmp_serde::to_vec(&my_map).unwrap();
    println!("Serialized bytes: {bytes:?}");
    let ptr: *const u8 = bytes.as_mut_ptr();
    compile_wasm(ptr, bytes.len());

    let x = alloc(128);
    dealloc(x, 128);
}
//fn main(x: Field, y: pub Field) {
//assert(x != y);

#[unsafe(no_mangle)]

pub extern "C" fn test_compile_wasm_go() {
    compile_from_memory(
        "fn main() {
    let mut acc: u128 = 0;
    let n: u128 = 1_000_000;

    for i in 0..n {
        acc = acc + (i * i) % 1234567;
        acc = acc ^ ((i + acc) % 987654321);
    }

    assert(acc == acc);
}",
    )
}
/*
#[unsafe(no_mangle)]
pub extern "C" fn SerializationTest(ptr: *const u8, len: usize) -> u8 {
    println!("made it into funciton Rust side");
    let data: &[u8] = unsafe { std::slice::from_raw_parts(ptr, len) };
    let map: HashMap<String, String> = rmp_serde::from_slice(data).expect("Serializing garbage");
    /*
        for (key, value) in map {
        }
    */

    let mut fm = file_manager_with_stdlib(Path::new(""));

    let mut crate_name = String::new();
    for (key, value) in map {
        println!("Key:{}|\nValue:{}|", key, value);

        if key == "CrateName" {
            crate_name = value;
            println!("{}<- crate name being passed in.", crate_name);
        } else {
            fm.add_file_with_source(Path::new(&key), value);
        }
    }
    let parsed_files = parse_all(&fm);

    let mut context = Context::new(fm, parsed_files);
    let options = CompileOptions::default();
    println!("Using main.nr");
    if crate_name == "" {
        panic!("Crate name is empty");
    }
    let crate_id = prepare_crate(&mut context, Path::new(&crate_name));
    let result = compile_main(&mut context, crate_id, &options, None);
    println!("boundary{:?}boundary", result);
    println!("Made it without crashing!");
    //panic!("Test Crash");
    1
}
*/

#[unsafe(no_mangle)]
pub extern "C" fn compile_wasm(ptr: *const u8, len: usize) /*-> (*const u8, usize)*/
{
    println!("made it into funciton Rust side");
    let data: &[u8] = unsafe { std::slice::from_raw_parts(ptr, len) };
    let map: HashMap<String, String> = rmp_serde::from_slice(data).expect("Serializing garbage");

    let mut fm = file_manager_with_stdlib(Path::new(""));

    let mut crate_name = String::new();
    for (key, value) in map {
        println!("Key:{}|\nValue:{}|", key, value);
        if key == "CrateName" {
            crate_name = value;
            println!("{}<- crate name being passed in.", crate_name);
        } else {
            fm.add_file_with_source(Path::new(&key), value);
        }
    }
    let parsed_files = parse_all(&fm);
    let mut context = Context::new(fm, parsed_files);
    let options = CompileOptions::default();
    println!("Using main.nr");
    if crate_name == "" {
        println!("Crate name is empty");
        //panic!("Crate name is empty");
    }
    println!("Crate name rust side |{}|", crate_name);
    //let crate_id = prepare_crate(&mut context, Path::new(&crate_name));
    //let result = compile_main(&mut context, crate_id, &options, None);
    //println!("boundary{:?}boundary", result);
    //println!("Made it without crashing!");
    //let compiled_program = result.expect("compilation failed").0;
    //let acir = compiled_program.program;
    //let bytes: Vec<u8> = Program::serialize_program(&acir);

    //println!("{:?}", bytes);
    //let aize = bytes.len();
    //let ptr = alloc(size);
    //unsafe {
    //ptr::copy_nonoverlapping(bytes.as_ptr(), ptr, size);
    //}
    //return (ptr, size);
    //let expression_width = acvm::acir::circuit::ExpressionWidth::Bounded { width: 4 };
    //let optimized_program = nargo::ops::transform_program(compiled_program, expression_width);

    //let program_artifact: noirc_artifacts::program::ProgramArtifact = optimized_program.into();
    //let bytecode_base64 = program_artifact.bytecode;
    //let acir_bytes = base64::decode(bytecode_base64).expect("Failed to decode base64 bytecode");

    // Now return acir_bytes as pointer and length
}

#[unsafe(no_mangle)]
pub extern "C" fn alloc(size: usize) -> *mut u8 {
    println!("{:?} <-- Rust size size ", size);
    let mut buf = Vec::with_capacity(size);
    let ptr = buf.as_mut_ptr();
    println!("{:?} <-- Rust size ptr ", ptr);
    std::mem::forget(buf);
    ptr
}

#[unsafe(no_mangle)]
pub extern "C" fn dealloc(ptr: *mut u8, size: usize) {
    unsafe {
        let _ = Vec::from_raw_parts(ptr, 0, size);
    }
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
