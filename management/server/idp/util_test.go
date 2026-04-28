package idp

import (
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	password, err := GeneratePassword(16, 2, 2, 2)
	if err != nil {
		t.Fatalf("GeneratePassword returned error: %v", err)
	}
	if len(password) != 16 {
		t.Fatalf("expected password length 16, got %d", len(password))
	}

	if countAny(password, specialCharSet) < 2 {
		t.Fatalf("expected at least 2 special characters in %q", password)
	}
	if countAny(password, numberSet) < 2 {
		t.Fatalf("expected at least 2 numeric characters in %q", password)
	}
	if countAny(password, upperCharSet) < 2 {
		t.Fatalf("expected at least 2 uppercase characters in %q", password)
	}
}

func countAny(value, charset string) int {
	count := 0
	for _, r := range value {
		if strings.ContainsRune(charset, r) {
			count++
		}
	}
	return count
}
