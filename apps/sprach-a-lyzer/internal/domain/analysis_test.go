package domain

import "testing"

func TestCanonicalDimension(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input DimensionID
		want  DimensionID
		ok    bool
	}{
		{input: DimensionVolition, want: DimensionVolition, ok: true},
		{input: "FREE_WILL", want: DimensionVolition, ok: true},
		{input: "UNKNOWN", want: "", ok: false},
	}

	for _, test := range tests {
		got, ok := CanonicalDimension(test.input)
		if got != test.want || ok != test.ok {
			t.Fatalf("CanonicalDimension(%q) = %q, %v; want %q, %v", test.input, got, ok, test.want, test.ok)
		}
	}
}
