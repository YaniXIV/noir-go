use acir::circuit::Program;
use acvm::acir::circuit::ExpressionWidth;
use base64::{Engine as _, engine::general_purpose};
use colored::Colorize;
use nargo::parse_all;
use noirc_abi::Abi;
use noirc_artifacts::program;
use noirc_driver::{CompileOptions, compile_main, file_manager_with_stdlib, prepare_crate};
use noirc_frontend::hir::Context;
use rmp_serde::encode::Serializer;
use rmp_serde::encode::to_vec;
use serde::{Deserialize, Serialize};
use serde_bytes::ByteBuf;
use std::collections::HashMap;
use std::panic::{self, AssertUnwindSafe};
use std::path::Path;
use std::ptr;

#[repr(C)]
pub struct WasmBuf {
    pub ptr: u32,
    pub len: u32,
}
#[derive(Serialize, Deserialize)]
struct MyMap(HashMap<String, String>);

#[derive(Serialize, Deserialize)]
struct WireCompileResult {
    pub format_version: u32,
    pub noir_version: String,
    pub abi_json: String,
    pub acir_string: String,
    pub acir_bytes: Vec<u8>,
    pub hash: u64,
    pub private_param_witnesses: Vec<u32>,
    pub public_param_witnesses: Vec<u32>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct WireMessage {
    /// Schema version for forward/back compat.
    #[serde(rename = "v")]
    pub version: u16,

    /// Message kind tag, e.g. "compile_ok", "compile_err".
    #[serde(rename = "k")]
    pub ok: bool,

    /// Msgpack bytes for the concrete payload struct.
    #[serde(with = "serde_bytes", rename = "p")]
    pub payload: Vec<u8>,
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
pub extern "C" fn compile_wasm(out: *mut WasmBuf, in_ptr: *mut u8, in_len: usize) {
    println!("made it into funciton Rust side");
    if in_ptr.is_null() || in_len == 0 {
        println!("{}", "Input buffer is empty.".red());
        unsafe {
            (*out).ptr = 0;
            (*out).len = 0;
        }
        return;
    }

    let data: &[u8] = unsafe { std::slice::from_raw_parts(in_ptr, in_len) };
    let map: HashMap<String, String> = match rmp_serde::from_slice(data) {
        Ok(map) => map,
        Err(err) => {
            println!("{} {err}", "Failed to deserialize input.".red());
            unsafe {
                (*out).ptr = 0;
                (*out).len = 0;
            }
            return;
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
            unsafe {
                (*out).ptr = 0;
                (*out).len = 0;
            }
            return;
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
                unsafe {
                    *out = WasmBuf { ptr: 0, len: 0 };
                }
                return;
            }
        },
        Err(_) => {
            eprintln!("Panic occured inside compiler");
            unsafe {
                *out = WasmBuf { ptr: 0, len: 0 };
            }
            return;
        }
    };
    let mut buf = Vec::new();

    wire.serialize(&mut Serializer::new(&mut buf).with_struct_map())
        .expect("msgpack serialization failed");
    let msgpack_bytes = buf;
    let out_len = msgpack_bytes.len();
    if out_len == 0 {
        unsafe {
            *out = WasmBuf { ptr: 0, len: 0 };
        }
        return;
    }
    let out_ptr = alloc(out_len);
    if out_ptr.is_null() {
        unsafe {
            *out = WasmBuf { ptr: 0, len: 0 };
        }
        return;
    }
    unsafe {
        ptr::copy_nonoverlapping(msgpack_bytes.as_ptr(), out_ptr, out_len);
    }
    unsafe {
        *out = WasmBuf {
            ptr: out_ptr as u32,
            len: out_len as u32,
        };
    }
    let json = serde_json::to_string_pretty(&wire).expect("failed to serialize");

    println!("{}", json);
    return;
}

fn compile_inner(
    context: &mut Context,
    crate_name: &str,
    options: &CompileOptions,
) -> Result<WireCompileResult, String> {
    let crate_id = prepare_crate(context, Path::new(crate_name));
    let result = compile_main(context, crate_id, options, None);

    match result {
        Ok((program, _)) => {
            let acir_program = &program.program;

            // Serialize ACIR
            let acir_bytes = Program::serialize_program(acir_program);

            let circuit = &acir_program.functions[0];

            // Extract private witness indices
            let private_indices: Vec<u32> = circuit
                .private_parameters
                .iter()
                .map(|w| w.witness_index())
                .collect();

            // Extract public witness indices
            let public_indices: Vec<u32> = circuit
                .public_parameters
                .0
                .iter()
                .map(|w| w.witness_index())
                .collect();

            Ok(WireCompileResult {
                format_version: 1,
                noir_version: program.noir_version,
                abi_json: serde_json::to_string(&program.abi).unwrap(),
                acir_string: acir_program.to_string(),
                acir_bytes,
                hash: program.hash,
                private_param_witnesses: private_indices,
                public_param_witnesses: public_indices,
            })
        }
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
