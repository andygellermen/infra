package pathutil

import "strings"

func Normalize(path string) string {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return "/"
	}
	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}
	if len(clean) == 1 {
		return clean
	}
	clean = strings.TrimRight(clean, "/")
	if clean == "" {
		return "/"
	}
	return clean
}
