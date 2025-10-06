package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("supersecretpassword")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "supersecretpassword"
	hash, _ := HashPassword(password)

	ok, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error: %v", err)
	}
	if !ok {
		t.Fatal("CheckPasswordHash failed for correct password")
	}

	ok, err = CheckPasswordHash("wrongpassword", hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash returned error for wrong password: %v", err)
	}
	if ok {
		t.Fatal("CheckPasswordHash returned true for wrong password")
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	secret := "secret"
	token, err := MakeJWT(userID, secret, time.Minute)
	if err != nil || token == "" {
		t.Fatal("MakeJWT failed")
	}
}

func TestValidateJWT(t *testing.T) {
	secret := "secret"
	userID := uuid.New()
	expiresIn := 1 * time.Second

	token, _ := MakeJWT(userID, secret, expiresIn)

	returnedID, err := ValidateJWT(token, secret)
	if err != nil || returnedID != userID {
		t.Fatal("ValidateJWT failed for valid token")
	}

	_, err = ValidateJWT(token, "wrongsecret")
	if err == nil {
		t.Fatal("ValidateJWT should fail with wrong secret")
	}

	time.Sleep(expiresIn)
	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("ValidateJWT should fail for expired token")
	}
}

func TestGetBearerToken(t *testing.T) {
	if _, err := GetBearerToken(http.Header{}); err == nil {
		t.Fatal("GetBearerToken should fail for headers without authorization header")
	}

	token, err := GetBearerToken(http.Header{"Authorization": []string{"Bearer mytoken"}})

	if err != nil {
		t.Fatalf("GetBearerToken returned error: %v", err)
	}

	if token != "mytoken" {
		t.Fatalf("GetBearerToken returned wrong token; expected token %s, got %s", token, "mytoken")
	}
}
