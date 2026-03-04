fn main( inputs: pub [Field; 4096], seed: Field) -> pub Field {
    let mut acc = seed;

    for i in 0..4096 {
        acc = acc * (inputs[i] + 1) + inputs[i];
    }

    assert(acc != 0);
    acc
}
