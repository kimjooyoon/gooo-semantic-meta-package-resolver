package resolver

import "testing"

func TestParseSourceRejectsDuplicateDeclarationKeys(t *testing.T) {
	_, err := ParseSourceBytes([]byte("gooo contract semantic_meta_package_resolver v1\ntype name=first definition=one name=second\n"))
	if err == nil {
		t.Fatal("expected duplicate declaration key to be rejected")
	}
}
