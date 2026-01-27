package auth

import (
	"testing"
)

func TestGenerateSecurePassword_Length(t *testing.T) {
	length := 16
	pw, err := GenerateSecurePassword(length)
	if err != nil {
		t.Fatalf("GenerateSecurePassword error: %v", err)
	}
	if len(pw) != length {
		t.Errorf("expected password length %d, got %d", length, len(pw))
	}
}

func TestGenerateSecurePassword_Unique(t *testing.T) {
	pw1, err1 := GenerateSecurePassword(12)
	pw2, err2 := GenerateSecurePassword(12)
	if err1 != nil || err2 != nil {
		t.Fatalf("error generating passwords: %v, %v", err1, err2)
	}
	if pw1 == pw2 {
		t.Error("expected different passwords, got the same")
	}
}
