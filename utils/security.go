/*
	Pseude security
	NOT a real communication authentication or authoraziton system implementation
	just a dummy procedure for pointing out that there must be an authentication or user validation system
*/
package utils

import "golang.org/x/crypto/bcrypt"

const (
	HashLength int = 60
)

var (
	secret []byte = []byte("secret")
)

func CreateHash() []byte {
	hash, _ := bcrypt.GenerateFromPassword(secret, bcrypt.DefaultCost)
	return hash
}

func ValidateHash(hash []byte) bool {
	return bcrypt.CompareHashAndPassword(hash, secret) == nil
}
