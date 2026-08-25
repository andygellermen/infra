package main

import "testing"

func TestResolveFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		requested string
		path      string
		want      string
	}{
		{requested: "auto", path: "seed.json", want: "json"},
		{requested: "auto", path: "seed.CSV", want: "csv"},
		{requested: "json", path: "without-extension", want: "json"},
	}
	for _, test := range tests {
		got, err := resolveFormat(test.requested, test.path)
		if err != nil || got != test.want {
			t.Fatalf("resolveFormat(%q, %q) = %q, %v; want %q", test.requested, test.path, got, err, test.want)
		}
	}
}
