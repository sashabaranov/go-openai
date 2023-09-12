package test

import (
	"crypto/rand"
	"math/big"
)

var (
	strLen = 10
	//nolint:gomnd // this avoids the golangci-lint "magic number" warning
	n = big.NewInt(2)
)

// RandomString generates a cryptographically secure random string of a fixed
// length. The string is composed of alphanumeric characters and is generated
// using the crypto/rand library. The length of the string is determined by the
// constant strLen.
func RandomString() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	s := make([]rune, strLen)
	max := big.NewInt(int64(len(letters)))
	for i := range s {
		randomInt, _ := rand.Int(rand.Reader, max)
		s[i] = letters[randomInt.Int64()]
	}
	return string(s)
}

// RandomBool generates a cryptographically secure random boolean value. It
// uses the crypto/rand library to generate a random integer (either 0 or 1),
// and returns true if the integer is 1, and false otherwise.
func RandomBool() bool {
	randomInt, _ := rand.Int(rand.Reader, n)
	return randomInt.Int64() == 1
}
