package auth

import "strings"

func EmailAllowed(allowed []string, candidate string) bool {
	email := strings.ToLower(strings.TrimSpace(candidate))
	if email == "" {
		return false
	}
	for _, item := range allowed {
		if strings.ToLower(strings.TrimSpace(item)) == email {
			return true
		}
	}
	return false
}
