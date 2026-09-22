package util

import (
	"testing"
	"time"
)

func TestGenerateAndParseToken(t *testing.T) {
	secret := "test-secret"
	token, err := GenerateToken(secret, time.Hour, 1, "alice", "DEVICE_ADMIN")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	claims, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken failed: %v", err)
	}
	if claims.UserID != 1 || claims.Username != "alice" || claims.Role != "DEVICE_ADMIN" {
		t.Errorf("claims mismatch: %+v", claims)
	}
}

func TestParseTokenInvalid(t *testing.T) {
	cases := []struct {
		name  string
		secret string
		token string
	}{
		{"wrong secret", "secret-a", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwibmFtZSI6ImEiLCJpYXQiOjE3MDAwMDAwMDB9.invalid"},
		{"empty token", "secret-a", ""},
		{"garbage", "secret-a", "not-a-jwt"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseToken(tc.secret, tc.token); err == nil {
				t.Errorf("expected error for token %q", tc.token)
			}
		})
	}
}
