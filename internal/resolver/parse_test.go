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

func TestLoadJSONRejectsTrailingValue(t *testing.T) {
	path := filepath.Join(t.TempDir(), "input.json")
	if err := os.WriteFile(path, []byte(`{"schema":"gooo/catalog/v1"} {}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var catalog Catalog
	if err := LoadJSON(path, &catalog); err == nil {
		t.Fatal("LoadJSON accepted a trailing JSON value")
	}
}
