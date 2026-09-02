package main

import "testing"

func TestInferFormat(t *testing.T) {
	t.Parallel()
	for input, want := range map[string]string{"data.JSON": "JSON", "data.csv": "CSV", "data.xlsx": "XLSX", "data.txt": ""} {
		if got := inferFormat(input); got != want {
			t.Errorf("inferFormat(%q)=%q want %q", input, got, want)
		}
	}
}
