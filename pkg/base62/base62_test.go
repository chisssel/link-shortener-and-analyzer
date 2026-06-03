package base62

import "testing"

func TestEncode_Zero(t *testing.T) {
	if got := Encode(0); got != "0" {
		t.Errorf("Encode(0) = %q, want %q", got, "0")
	}
}

func TestEncode_One(t *testing.T) {
	if got := Encode(1); got != "1" {
		t.Errorf("Encode(1) = %q, want %q", got, "1")
	}
}

func TestEncode_Ten(t *testing.T) {
	if got := Encode(10); got != "A" {
		t.Errorf("Encode(10) = %q, want %q", got, "A")
	}
}

func TestEncode_SixtyOne(t *testing.T) {
	if got := Encode(61); got != "z" {
		t.Errorf("Encode(61) = %q, want %q", got, "z")
	}
}

func TestEncode_SixtyTwo(t *testing.T) {
	if got := Encode(62); got != "10" {
		t.Errorf("Encode(62) = %q, want %q", got, "10")
	}
}

func TestEncode_Large(t *testing.T) {
	if got := Encode(123456); got != "W7E" {
		t.Errorf("Encode(123456) = %q, want %q", got, "W7E")
	}
}

func TestEncode_MaxUint64(t *testing.T) {
	got := Encode(18446744073709551615)
	if len(got) == 0 || got[0] != 'L' {
		t.Errorf("Encode(max) seems wrong: %q", got)
	}
}

func TestDecode_Zero(t *testing.T) {
	if got := Decode("0"); got != 0 {
		t.Errorf("Decode(\"0\") = %d, want %d", got, 0)
	}
}

func TestDecode_Ten(t *testing.T) {
	if got := Decode("A"); got != 10 {
		t.Errorf("Decode(\"A\") = %d, want %d", got, 10)
	}
}

func TestDecode_SixtyTwo(t *testing.T) {
	if got := Decode("10"); got != 62 {
		t.Errorf("Decode(\"10\") = %d, want %d", got, 62)
	}
}

func TestDecode_InvalidChar(t *testing.T) {
	// '!' is not in the alphabet
	if got := Decode("!"); got != 0 {
		t.Errorf("Decode(\"!\") = %d, want %d", got, 0)
	}
}

func TestDecode_MixedCase(t *testing.T) {
	a := Decode("abc")
	b := Decode("ABC")
	if a == b {
		t.Errorf("Decode should be case-sensitive: abc=%d ABC=%d", a, b)
	}
}

func TestRoundTrip_Small(t *testing.T) {
	ids := []uint64{0, 1, 5, 10, 42, 61, 62, 100, 999}
	for _, id := range ids {
		encoded := Encode(id)
		decoded := Decode(encoded)
		if decoded != id {
			t.Errorf("RoundTrip(%d) = %q -> %d", id, encoded, decoded)
		}
	}
}

func TestRoundTrip_Large(t *testing.T) {
	ids := []uint64{100000, 123456789, 9999999999, 18446744073709551615}
	for _, id := range ids {
		encoded := Encode(id)
		decoded := Decode(encoded)
		if decoded != id {
			t.Errorf("RoundTrip(%d) = %q -> %d", id, encoded, decoded)
		}
	}
}

func TestEncode_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := uint64(0); i < 10000; i++ {
		code := Encode(i)
		if seen[code] {
			t.Errorf("collision at %d: %q", i, code)
		}
		seen[code] = true
	}
}

func TestEncodeDecode_Random(t *testing.T) {
	for i := 0; i < 100; i++ {
		id := uint64(i * 100000)
		encoded := Encode(id)
		decoded := Decode(encoded)
		if decoded != id {
			t.Errorf("RoundTrip(%d) failed: %q -> %d", id, encoded, decoded)
		}
	}
}
