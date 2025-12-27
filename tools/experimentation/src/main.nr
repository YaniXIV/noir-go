#Stress test for noir compiler 

fn main() {
    let mut acc: u128 = 0;
    let n: u128 = 1_000_000;

    for i in 0..n {
        acc = acc + (i * i) % 1234567;
        acc = acc ^ ((i + acc) % 987654321);
    }

    assert(acc == acc);
}
