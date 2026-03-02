fn main(x: Field, y: Field) -> pub Field {
    let mut acc = x;

    for _i in 0..20000{
        acc = acc * y + x;
        acc = acc * acc + y;
        acc = acc - x;
        acc = acc * (y + 1);
    }

    assert(acc != 0);
    acc
}
