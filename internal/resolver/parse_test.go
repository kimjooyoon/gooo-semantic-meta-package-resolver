package resolver

import "testing"

func TestParseKeyValuesRejectsDuplicateKeys(t *testing.T) {
	if _, err := parseKeyValues([]string{"name=first", "name=second"}); err == nil {
		t.Fatal("parseKeyValues accepted duplicate keys")
	}
}

func TestStripCommentPreservesQuotedHash(t *testing.T) {
	line := `export id="pkg#alias" # trailing comment`
	if got, want := stripComment(line), `export id="pkg#alias" `; got != want {
		t.Fatalf("stripComment(%q)=%q want %q", line, got, want)
	}
}
