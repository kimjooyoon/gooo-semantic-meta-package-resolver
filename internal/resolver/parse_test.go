package resolver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseSourceRejectsDuplicateDeclarationKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.gooo")
	if err := os.WriteFile(path, []byte("gooo contract semantic_meta_package_resolver v1\ntype name=first definition=one name=second\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ParseSource(path)
	if err == nil {
		t.Fatal("expected duplicate declaration key to be rejected")
	}
}

func TestParseSourceRejectsDuplicateHeaders(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.gooo")
	if err := os.WriteFile(path, []byte("gooo contract semantic_meta_package_resolver v1\ngooo consumer example v2\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := ParseSource(path)
	if err == nil {
		t.Fatal("expected duplicate gooo header to be rejected")
	}
}

func TestParseSourceRejectsDuplicatePackageMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.gooo")
	data := "gooo contract semantic_meta_package_resolver v1\npackage name=first\npackage name=second\n"
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseSource(path); err == nil {
		t.Fatal("expected duplicate package metadata to be rejected")
	}
}
