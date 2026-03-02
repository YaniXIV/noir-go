fn main(x: Field, y: Field) -> pub Field {
    let z = x * y + x;
    assert(z != 0);
    z
}
