use acir::AcirField;
use acir::circuit::Program;
use acvm::{
    FieldElement,
    acir::native_types::{Witness, WitnessMap},
    pwg::{ACVM, ACVMStatus},
};
use bn254_blackbox_solver::Bn254BlackBoxSolver;
use colored::Colorize;
use nargo::parse_all;
use noirc_driver::{CompileOptions, compile_main, file_manager_with_stdlib, prepare_crate};
use noirc_frontend::hir::Context;
use rmp_serde::encode::Serializer;
use serde::{Deserialize, Serialize};
use std::any::Any;
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
    pub acir_json: String,
    pub hash: u64,
    pub private_param_witnesses: Vec<u32>,
    pub public_param_witnesses: Vec<u32>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct WireMessage {
    /// Schema version for forward/back compat.
    #[serde(rename = "v")]
    pub version: u16,

    /// True if `payload` decodes to the success struct; false if it decodes
    /// to a WireError.
    #[serde(rename = "k")]
    pub ok: bool,

    /// Msgpack bytes for the concrete payload struct.
    #[serde(with = "serde_bytes", rename = "p")]
    pub payload: Vec<u8>,
}

/// Error payload carried inside a WireMessage when `ok == false`.
#[derive(Serialize, Deserialize)]
struct WireError {
    pub message: String,
}

/// Encodes `payload` (already-msgpack-encoded bytes of either the success
/// struct or a WireError) into a WireMessage envelope, allocates guest
/// memory for it, and writes the resulting pointer/length into `out`.
///
/// Every exit path of compile_wasm/execute_wasm goes through this (via
/// write_error/write_success below) so a Go caller always gets a valid,
/// decodable buffer -- including on failure. `zero_buf` is reserved solely
/// for the case where we can't even allocate/serialize the envelope itself.
fn zero_buf(out: *mut WasmBuf) {
    unsafe {
        *out = WasmBuf { ptr: 0, len: 0 };
    }
}

fn write_envelope(out: *mut WasmBuf, ok: bool, payload: Vec<u8>) {
    let msg = WireMessage {
        version: 1,
        ok,
        payload,
    };
    let mut buf = Vec::new();
    if msg
        .serialize(&mut Serializer::new(&mut buf).with_struct_map())
        .is_err()
    {
        zero_buf(out);
        return;
    }

    let out_len = buf.len();
    if out_len == 0 {
        zero_buf(out);
        return;
    }
    let out_ptr = alloc(out_len);
    if out_ptr.is_null() {
        zero_buf(out);
        return;
    }
    unsafe {
        ptr::copy_nonoverlapping(buf.as_ptr(), out_ptr, out_len);
        *out = WasmBuf {
            ptr: out_ptr as u32,
            len: out_len as u32,
        };
    }
}

fn write_error(out: *mut WasmBuf, message: String) {
    // The message is already carried structurally in the WireError payload,
    // so it isn't also echoed to stderr here -- stderr capture on the Go
    // side is for genuinely separate diagnostics (e.g. Rust's own panic
    // hook output), not a duplicate of this string.
    let err = WireError { message };
    let mut payload = Vec::new();
    match err.serialize(&mut Serializer::new(&mut payload).with_struct_map()) {
        Ok(()) => write_envelope(out, false, payload),
        Err(_) => zero_buf(out),
    }
}

fn write_success<T: Serialize>(out: *mut WasmBuf, value: &T) {
    let mut payload = Vec::new();
    match value.serialize(&mut Serializer::new(&mut payload).with_struct_map()) {
        Ok(()) => write_envelope(out, true, payload),
        Err(e) => write_error(out, format!("failed to serialize result: {e}")),
    }
}

/// Best-effort extraction of a message from a caught panic payload.
fn panic_message(payload: Box<dyn Any + Send>) -> String {
    if let Some(s) = payload.downcast_ref::<&str>() {
        s.to_string()
    } else if let Some(s) = payload.downcast_ref::<String>() {
        s.clone()
    } else {
        "panic occurred inside compiler (no message)".to_string()
    }
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
    let _ptr: *const u8 = bytes.as_mut_ptr();

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
        write_error(out, "Input buffer is empty.".to_string());
        return;
    }

    let data: &[u8] = unsafe { std::slice::from_raw_parts(in_ptr, in_len) };
    let map: HashMap<String, String> = match rmp_serde::from_slice(data) {
        Ok(map) => map,
        Err(err) => {
            write_error(out, format!("Failed to deserialize input: {err}"));
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
            write_error(out, "Crate name is empty".to_string());
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

    // Panics are caught here so a broken/unexpected input can never trap the
    // whole WASM instance -- it always comes back as a normal Go error with
    // whatever message we could extract from the panic payload instead.
    let wire = match caught {
        Ok(inner) => match inner {
            Ok(wire) => wire,
            Err(msg) => {
                write_error(out, msg);
                return;
            }
        },
        Err(panic_payload) => {
            write_error(
                out,
                format!(
                    "Panic occurred inside compiler: {}",
                    panic_message(panic_payload)
                ),
            );
            return;
        }
    };

    let json = serde_json::to_string_pretty(&wire).unwrap_or_default();
    println!("{}", json);

    write_success(out, &wire);
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
            let acir_json = serde_json::to_string(acir_program).unwrap();

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
                acir_json: acir_json,
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

#[derive(Serialize, Deserialize)]
struct WireExecuteInput {
    pub acir_bytes: Vec<u8>,
    pub initial_witness: Vec<(u32, [u8; 32])>,
}

#[derive(Serialize, Deserialize)]
struct WireExecuteResult {
    pub witness: Vec<(u32, [u8; 32])>,
}

fn execute_inner(data: &[u8]) -> Result<WireExecuteResult, String> {
    let input: WireExecuteInput = rmp_serde::from_slice(data)
        .map_err(|e| format!("execute_wasm: deserialize input failed: {e}"))?;

    let program: acvm::acir::circuit::Program<FieldElement> =
        acvm::acir::circuit::Program::deserialize_program(&input.acir_bytes)
            .map_err(|e| format!("execute_wasm: deserialize program failed: {e:?}"))?;

    let circuit = &program.functions[0];

    let mut initial_witness = WitnessMap::new();
    for (idx, bytes) in input.initial_witness {
        let fe = FieldElement::from_le_bytes_reduce(&bytes);
        initial_witness.insert(Witness(idx), fe);
    }

    let solver = Bn254BlackBoxSolver(false);
    let mut acvm = ACVM::new(
        &solver,
        &circuit.opcodes,
        initial_witness,
        &program.unconstrained_functions,
        &circuit.assert_messages,
    );

    loop {
        match acvm.solve() {
            ACVMStatus::Solved => break,
            ACVMStatus::InProgress => continue,
            ACVMStatus::Failure(e) => {
                return Err(format!("execute_wasm: solver failure: {e}"));
            }
            ACVMStatus::RequiresForeignCall(_) => {
                use acvm::brillig_vm::brillig::ForeignCallResult;
                acvm.resolve_pending_foreign_call(ForeignCallResult::default());
            }
            ACVMStatus::RequiresAcirCall(_) => {
                return Err("execute_wasm: ACIR calls not supported yet".to_string());
            }
        }
    }

    let witness_map = acvm.finalize();
    let mut witness_out = Vec::new();
    for (w, fe) in witness_map.into_iter() {
        let bytes: [u8; 32] = fe
            .to_be_bytes()
            .try_into()
            .map_err(|_| "execute_wasm: witness value was not 32 bytes".to_string())?;
        witness_out.push((w.witness_index(), bytes));
    }

    Ok(WireExecuteResult {
        witness: witness_out,
    })
}

#[unsafe(no_mangle)]
pub extern "C" fn execute_wasm(out: *mut WasmBuf, in_ptr: *mut u8, in_len: usize) {
    if in_ptr.is_null() || in_len == 0 {
        write_error(out, "execute_wasm: input buffer is empty".to_string());
        return;
    }
    let data: &[u8] = unsafe { std::slice::from_raw_parts(in_ptr, in_len) };

    // Panics (e.g. from an unexpected witness/program shape) are caught here
    // for the same reason as in compile_wasm: never trap the WASM instance,
    // always come back as a normal Go error.
    let caught = panic::catch_unwind(AssertUnwindSafe(|| execute_inner(data)));

    let wire = match caught {
        Ok(Ok(wire)) => wire,
        Ok(Err(msg)) => {
            write_error(out, msg);
            return;
        }
        Err(panic_payload) => {
            write_error(
                out,
                format!(
                    "Panic occurred inside solver: {}",
                    panic_message(panic_payload)
                ),
            );
            return;
        }
    };

    write_success(out, &wire);
}
