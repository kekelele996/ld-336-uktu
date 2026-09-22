package util

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("medasset123")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "medasset123" {
		t.Error("password should be hashed")
	}
	if !CheckPassword(hash, "medasset123") {
		t.Error("CheckPassword should return true for correct password")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("CheckPassword should return false for wrong password")
	}
}
