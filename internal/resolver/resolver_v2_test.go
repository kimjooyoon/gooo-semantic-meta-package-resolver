package resolver

import (
	"path/filepath"
	"testing"
)

func TestV2FixedVector(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	result, err := RunV2Conformance(root, filepath.Join(root, "fixtures"), t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if result.Cases != 12 || result.Closed != 4 || result.Unknown != 4 || result.Refuted != 4 || result.Cells != 12 || result.MetaActivities != 12 {
		t.Fatalf("unexpected v2 conformance summary: %+v", result)
	}
}

func TestV2IdentityTupleAndPrecedence(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	contract, err := ParseSource(filepath.Join(root, "contracts", "resolver-v2.gooo"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateV2Contract(contract); err != nil {
		t.Fatal(err)
	}
	status, claim := finalClaim([]Claim{
		unknownClaim("IDENTITY", "VERIFY_ASSET", "ASSET_DIGEST_UNAVAILABLE", "MISSING_ASSET_DIGEST", "OBTAIN_ASSET_DIGEST", "asset"),
		refutedClaim("IDENTITY", "VERIFY_PACKAGE", "FORGED_PACKAGE_IDENTITY", "package"),
	})
	if status != Refuted || claim.Reason != "FORGED_PACKAGE_IDENTITY" {
		t.Fatalf("v2 precedence was not REFUTED > UNKNOWN: %s %+v", status, claim)
	}
}
