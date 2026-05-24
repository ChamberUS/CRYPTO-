package types

import "testing"

func TestBaseDenomAndSupplyConstants(t *testing.T) {
	if BaseDenom != "ubyx" {
		t.Fatalf("BaseDenom mismatch: got %q want %q", BaseDenom, "ubyx")
	}
	if DisplayDenom != "BYX" {
		t.Fatalf("DisplayDenom mismatch: got %q want %q", DisplayDenom, "BYX")
	}
	if DenomBYX != BaseDenom {
		t.Fatalf("DenomBYX should alias BaseDenom")
	}
	if UbyxPerByx != 1_000_000 {
		t.Fatalf("UbyxPerByx mismatch: got %d", UbyxPerByx)
	}
	if MaxSupplyUbyx != 1_000_000_000_000_000 {
		t.Fatalf("MaxSupplyUbyx mismatch: got %d", MaxSupplyUbyx)
	}
}
