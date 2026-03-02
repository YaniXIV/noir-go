fn main(a: u32, b: u32) -> pub u32 {
    let mut x = a + 1;
    let mut y = b + 1;

    // Keep numbers bounded to avoid overflow issues on older Noir versions.
    // Using a modulus also forces more nontrivial integer constraints.
    for i in 0..5000 {
        // "Mix" step: multiply + add
        x = (x * 1664525 + 1013904223) % 4294967291; // prime-ish < 2^32
        y = (y * 22695477 + 1) % 4294967291;

        // Add some extra work (division/mod are expensive-ish in many constraint systems)
        let d = (x % 97) + 1;      // avoid div by 0
        let q = y / d;
        let r = y - q * d;         // remainder relationship

        // Fold back in so the compiler can't simplify it away
        x = (x + q + r + (i as u32)) % 4294967291;
        y = (y + x + 3) % 4294967291;
    }

    assert(x != 0);
    x
}
