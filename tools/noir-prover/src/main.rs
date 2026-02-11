fn main() {
    println!("FOOBAR");
}

#[repr(C)]
pub struct Slice {
    ptr: *mut u8,
    len: usize,
}

/*
* For later
*
* pub fn generate_partial_witness(acir: &AcirInstance, initial_witness: &WitnessMap, foreign_call_handler: &impl ForeignCallHandler) -> Result<WitnessMap, AcvmError>
*
* check acvm crates pwg module
*
*
*/

#[unsafe(no_mangle)]
pub extern "C" fn generate_witness(
    acir_ptr: *const u8,
    acir_len: usize,
    in_ptr: *const u8,
    in_len: usize,
    out_ptr: *mut Slice,
) {
    let acir_bytes = unsafe { std::slice::from_raw_parts(acir_ptr, acir_len) };
    let in_bytes = unsafe { std::slice::from_raw_parts(in_ptr, in_len) };
    /*
     * START WORKING ON WINESS GENERATION
     *
     * STEP 1, COLLECT INPUTS, serialize initial witness into WitnessMap
     * step 2, use acir bytes and intial witnese to feed acvm PWG
     * step 3, get full witness and move on to bb.
     *
     * not sure if I want to support multiple proving backends yet, I feel like
     * I could do it. with acir,initial witness. gonna start with bb first
     * perhpalps setup config for prover loading so you can choose other
     * proving backends to plug and play :) we'll see tho.
     *
     */
}
