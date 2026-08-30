package service

import "testing"

func TestGenerateAndParseToken_RoundTrip(t *testing.T) {
	s := &AuthService{jwtSecret: "test-secret"}

	token, err := s.generateToken(42)
	if err != nil {
		t.Fatalf("generateToken() error = %v", err)
	}

	userID, err := s.ParseToken(token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

	if userID != 42 {
		t.Errorf("ParseToken() userID = %d, want 42", userID)
	}
}

func TestParseToken_WrongSecretIsRejected(t *testing.T) {
	signer := &AuthService{jwtSecret: "secret-a"}
	verifier := &AuthService{jwtSecret: "secret-b"}

	token, err := signer.generateToken(1)
	if err != nil {
		t.Fatalf("generateToken() error = %v", err)
	}

	if _, err := verifier.ParseToken(token); err == nil {
		t.Error("ParseToken() expected error for token signed with a different secret, got nil")
	}
}

func TestParseToken_GarbageIsRejected(t *testing.T) {
	s := &AuthService{jwtSecret: "test-secret"}

	if _, err := s.ParseToken("not-a-real-token"); err == nil {
		t.Error("ParseToken() expected error for malformed token, got nil")
	}
}
