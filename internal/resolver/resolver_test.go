package resolver

import (
	"path/filepath"
	"testing"
)

func TestFixedSevenCases(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	result, err := RunConformance(root, filepath.Join(root, "fixtures"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if result.Cases != 7 || result.Closed != 4 || result.Unknown != 1 || result.Refuted != 2 {
		t.Fatalf("unexpected conformance summary: %+v", result)
	}
}

func TestUnknownClaimIsComplete(t *testing.T) {
	claim := unknownClaim("IMMUTABLE_PROOF", "VERIFY_CATALOG_PROOF", "IMMUTABLE_PROOF_UNAVAILABLE", "MISSING_IMMUTABLE_PROOF", "OBTAIN_IMMUTABLE_RELEASE_PROOF", "unproven")
	if !claim.Valid() {
		t.Fatalf("unknown claim did not satisfy six-field contract: %+v", claim)
	}
}

func TestPrecedence(t *testing.T) {
	status, claim := finalClaim([]Claim{
		unknownClaim("A", "B", "C", "D", "E", "unknown"),
		refutedClaim("A", "B", "Z", "refuted"),
	})
	if status != Refuted || claim.Reason != "Z" {
		t.Fatalf("precedence was not REFUTED > UNKNOWN: %s %+v", status, claim)
	}
}

func TestCaretConstraintsRespectZeroMajorSemver(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		constraint string
		want       bool
	}{
		{name: "zero-minor-inside", version: "0.2.9", constraint: "^0.2.3", want: true},
		{name: "zero-minor-upper-bound", version: "0.3.0", constraint: "^0.2.3", want: false},
		{name: "zero-minor-wide-version", version: "0.99.0", constraint: "^0.2.3", want: false},
		{name: "zero-patch-inside", version: "0.0.3", constraint: "^0.0.3", want: true},
		{name: "zero-patch-upper-bound", version: "0.0.4", constraint: "^0.0.3", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := satisfies(tt.version, tt.constraint); got != tt.want {
				t.Fatalf("satisfies(%q, %q)=%v want %v", tt.version, tt.constraint, got, tt.want)
			}
		})
	}
}
