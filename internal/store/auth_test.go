package store

import "testing"

func TestPasswordHashAndAccessToken(t *testing.T) {
	hash, err := hashPassword("admin123")
	if err != nil {
		t.Fatal(err)
	}
	if !checkPassword(hash, "admin123") {
		t.Fatal("expected password to match")
	}
	if checkPassword(hash, "wrong") {
		t.Fatal("expected wrong password to fail")
	}

	token, expiresIn, err := newAccessToken(42)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || expiresIn <= 0 {
		t.Fatalf("invalid token result: token=%q expires=%d", token, expiresIn)
	}
	userID, err := parseAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if userID != 42 {
		t.Fatalf("expected user 42, got %d", userID)
	}
}

func TestRefreshTokenHash(t *testing.T) {
	token, hash, expiresAt, err := newRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || hash == "" || expiresAt.IsZero() {
		t.Fatalf("invalid refresh token result")
	}
	if tokenHash(token) != hash {
		t.Fatal("refresh token hash mismatch")
	}
}
