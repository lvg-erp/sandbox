package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestTokenRoundtrip(t *testing.T) {
	cfg := NewJWTConfig("test-secret")
	id := uuid.New()

	token, err := cfg.GenerateAccessToken(id, "alice")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	claims, err := cfg.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserUUID != id.String() {
		t.Fatalf("unexpected user UUID: %s", claims.UserUUID)
	}
	if claims.Username != "alice" {
		t.Fatalf("unexpected username: %s", claims.Username)
	}
}

func TestValidateWrongSecret(t *testing.T) {
	cfg := NewJWTConfig("test-secret")
	token, err := cfg.GenerateAccessToken(uuid.New(), "alice")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	other := NewJWTConfig("other-secret")
	if _, err := other.ValidateToken(token); err == nil {
		t.Fatal("expected error for token signed with another secret")
	}
}

func TestValidateGarbage(t *testing.T) {
	cfg := NewJWTConfig("test-secret")
	if _, err := cfg.ValidateToken("not-a-token"); err == nil {
		t.Fatal("expected error for garbage token")
	}
}

func TestGenerateRandomToken(t *testing.T) {
	a, err := GenerateRandomToken()
	if err != nil {
		t.Fatalf("GenerateRandomToken: %v", err)
	}
	b, err := GenerateRandomToken()
	if err != nil {
		t.Fatalf("GenerateRandomToken: %v", err)
	}

	// 32 байта в hex
	if len(a) != 64 {
		t.Fatalf("unexpected token length: %d", len(a))
	}
	if a == b {
		t.Fatal("tokens must be unique")
	}
}
