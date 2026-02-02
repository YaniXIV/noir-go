use acir::circuit::Program;
use acvm::acir::circuit::ExpressionWidth;
use colored::Colorize;
use nargo::parse_all;
use noirc_abi::Abi;
use noirc_driver::{CompileOptions, compile_main, file_manager_with_stdlib, prepare_crate};
use noirc_frontend::hir::Context;
use rmp_serde::encode::to_vec;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::panic::{self, AssertUnwindSafe};
use std::path::Path;
use std::ptr;

#[derive(Serialize, Deserialize)]
struct MyMap(HashMap<String, String>);
#[derive(Serialize, Deserialize)]
struct WireCompileResult {
    pub format_version: u32,
    pub noir_version: String,
    pub abi: Abi,
    pub acir_string: String,
    pub acir_bytes: Vec<u8>,
    pub hash: u64,
}

fn main() {
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
    //compile_wasm(ptr, bytes.len());

    let x = alloc(128);
    dealloc(x, 128);
}

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

#[unsafe(no_mangle)]
pub extern "C" fn compile_wasm(in_ptr: *mut u8, in_len: usize) -> (*const u8, usize) {
    println!("made it into funciton Rust side");
    if in_ptr.is_null() || in_len == 0 {
        println!("{}", "Input buffer is empty.".red());
        return (ptr::null(), 0);
    }

    let data: &[u8] = unsafe { std::slice::from_raw_parts(in_ptr, in_len) };
    let map: HashMap<String, String> = match rmp_serde::from_slice(data) {
        Ok(map) => map,
        Err(err) => {
            println!("{} {err}", "Failed to deserialize input.".red());
            return (ptr::null(), 0);
        }
    };

    let mut fm = file_manager_with_stdlib(Path::new(""));

    let mut crate_name = None;
    for (key, value) in map {
        println!("{}|{}|", "Key:".red(), key);
        println!("{}|{}|", "Value:".blue().bold(), value);

        if key == "CrateName" {
            println!("{}|{}|", "Crate name being passed in.".green(), value);
            crate_name = Some(value);
        } else {
            println!("\nPassing in |{}|\n\nand\n\n|{}|", &key, value);
            fm.add_file_with_source(Path::new(&key), value);
        }
    }
    let parsed_files = parse_all(&fm);

    let mut context = Context::new(fm, parsed_files);
    let options = CompileOptions::default();
    println!("{}", "Using main.nr".yellow());
    let crate_name = match crate_name {
        Some(name) => name,
        None => {
            println!("{}", "Crate name is empty".red());
            return (ptr::null(), 0);
        }
    };
    let testpath = "/Users/yani/noir-go/internal/compiler/noirtest/src/main.nr";
    if crate_name != testpath {
        println!("crate_name does not match testpath");
        println!(
            "\nPassing in |{}|\n\ntestpath is\n\n|{}|",
            crate_name, testpath
        );
    }
    let caught = panic::catch_unwind(AssertUnwindSafe(|| {
        compile_inner(&mut context, &crate_name, &options)
    }));
    //Hopefully lets us catch panics.
    //Although ive been having an issue before where panics were not caught.
    //Could be a wasm rust thing, like how this specific panic in
    //prepare_crate throws and how it unwinds.
    //Ill maybe just panic later to test it.
    let wire = match caught {
        Ok(inner) => match inner {
            Ok(wire) => wire,
            Err(msg) => {
                eprintln!("{msg}");
                return (ptr::null(), 0);
            }
        },
        Err(_) => {
            eprintln!("Panic occured inside compiler");
            return (ptr::null(), 0);
        }
    };
    let msgpack_bytes: Vec<u8> = to_vec(&wire).expect("msgpack serialization failed");
    let out_len = msgpack_bytes.len();
    if out_len == 0 {
        return (std::ptr::null(), 0);
    }
    let out_ptr = alloc(out_len);
    if out_ptr.is_null() {
        return (std::ptr::null(), 0);
    }
    unsafe {
        ptr::copy_nonoverlapping(msgpack_bytes.as_ptr(), out_ptr, out_len);
    }
    return (out_ptr, out_len);
}

fn compile_inner(
    context: &mut Context,
    crate_name: &str,
    options: &CompileOptions,
) -> Result<WireCompileResult, String> {
    let crate_id = prepare_crate(context, Path::new(&crate_name));
    let result = compile_main(context, crate_id, &options, None);
    match result {
        Ok((program, _)) => Ok(WireCompileResult {
            format_version: 1,
            noir_version: program.noir_version,
            abi: program.abi,
            acir_string: program.program.to_string(),
            acir_bytes: Program::serialize_program(&program.program),
            hash: program.hash,
        }),
        Err(err) => Err(format!("Compilation failed: {err:?}")),
    }
}

#[unsafe(no_mangle)]
pub extern "C" fn alloc(size: usize) -> *mut u8 {
    //println!("{:?} <-- Rust size size ", size);
    let mut buf = Vec::with_capacity(size);
    let ptr = buf.as_mut_ptr();
    //println!("{:?} <-- Rust size ptr ", ptr);
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
