package persist

import (
	"log"
	// "unsafe"
	
	"golang.org/x/crypto/bcrypt"
)

// EncryptPassword will encrypt the password and replace the old one. Will return the encrypted password.
func (user *User) EncryptPassword() string {
	if user.Password == "" {
		log.Println("Nothing to encrypt.")

		return ""
	}

	hpass, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
	if err != nil {
		log.Fatal(err)
	}

	user.Password = string(hpass)

	return user.Password
}

// CheckPassword returns true if passwords match.
func (user *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))

	return err == nil
}