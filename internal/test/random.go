package test

import (
	"crypto/rand"
	"strings"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const strLen = 10

// See StackOverflow answer:
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
// RandomString generates a cryptographically secure random string of length
// strLen.
func RandomString() string {
	sb := strings.Builder{}
	sb.Grow(strLen)

	for i := 0; i < strLen; i++ {
		randomByte := make([]byte, 1)
		_, err := rand.Read(randomByte)
		if err != nil {
			return ""
		}
		randomIndex := randomByte[0] % byte(len(letters))
		sb.WriteByte(letters[randomIndex])
	}

	return sb.String()
}

// RandomBool generates a cryptographically secure random boolean value.
// It reads a single byte from the crypto/rand library and uses its least
// significant bit to determine the boolean value. The function returns
// true if the least significant bit is 1, and false otherwise.
func RandomBool() bool {
	var b [1]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return false
	}
	return b[0]&1 == 1
}
