package resolver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseKeyValuesRejectsDuplicateKeys(t *testing.T) {
	if _, err := parseKeyValues([]string{"name=first", "name=second"}); err == nil {
		t.Fatal("parseKeyValues accepted duplicate keys")
	}
}

func TestParseSourceRejectsDuplicateGoooHeaders(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.gooo")
	content := "gooo consumer demo v1\n" +
		"gooo consumer replacement v1\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseSource(path); err == nil {
		t.Fatal("ParseSource accepted duplicate gooo headers")
	}
}
