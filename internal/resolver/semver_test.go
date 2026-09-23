package resolver

import "testing"

func TestCaretConstraintUsesSemverZeroMajorBounds(t *testing.T) {
	tests := []struct {
		version    string
		constraint string
		want       bool
	}{
		{version: "0.2.4", constraint: "^0.2.3", want: true},
		{version: "0.9.0", constraint: "^0.2.3", want: false},
		{version: "0.0.3", constraint: "^0.0.3", want: true},
		{version: "0.0.4", constraint: "^0.0.3", want: false},
	}
	for _, test := range tests {
		if got := satisfies(test.version, test.constraint); got != test.want {
			t.Fatalf("satisfies(%q, %q) = %t, want %t", test.version, test.constraint, got, test.want)
		}
	}
}
