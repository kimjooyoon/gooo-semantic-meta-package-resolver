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

func TestParseSourcePreservesHashesInsideQuotedValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.gooo")
	content := `gooo syntax sample v1
grammar value="grammar#1"
rule name="rule#1" definition="definition#1"
semantic_id value="semantic#1"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	source, err := ParseSource(path)
	if err != nil {
		t.Fatalf("ParseSource rejected hashes inside quoted values: %v", err)
	}
	if len(source.Grammar) != 1 || source.Grammar[0] != "grammar#1" {
		t.Fatalf("grammar quoted value lost hash: %+v", source.Grammar)
	}
	if len(source.Rules) != 1 || source.Rules[0] != "rule#1=definition#1" {
		t.Fatalf("rule quoted values lost hashes: %+v", source.Rules)
	}
	if source.SemanticID != "semantic#1" {
		t.Fatalf("semantic ID lost hash: %q", source.SemanticID)
	}
}
