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
