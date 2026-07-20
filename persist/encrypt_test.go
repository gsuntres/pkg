package persist

import (
	"testing"
)

func TestUser_EncryptPassword(t *testing.T) {
	u := &User{
		Password: "1234",
	}

	encrypted := u.EncryptPassword()

	if u.Password != encrypted {
		t.Fatal("Failed to encrypt password")
	}
}

func TestUser_CheckPassword(t *testing.T) {
	u := &User{
		Password: "1234",
	}

	u.EncryptPassword()

	if !u.CheckPassword("1234") {
		t.Fatal("Failed to check password")
	}
}

func TestUser_CheckPassword_FalsePositive(t *testing.T) {
	u := &User{
		Password: "1234",
	}

	u.EncryptPassword()

	if u.CheckPassword("!234") {
		t.Fatal("Should have not accepted password")
	}
}