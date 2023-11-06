package test

import (
	"math/rand"
	"strings"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const strLen = 10

// #nosec G404
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// Seeding func.
// #nosec G404
func Seed(s int64) {
	r = rand.New(rand.NewSource(s))
}

// See StackOverflow answer:
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
// RandomString generates a random string of length
// strLen.
func RandomString() string {
	sb := strings.Builder{}
	sb.Grow(strLen)

	for i := 0; i < strLen; i++ {
		randomIndex := r.Intn(len(letters))
		sb.WriteByte(letters[randomIndex])
	}

	return sb.String()
}

// RandomInt generates a random integer between 0 (inclusive) and 'max'
// (exclusive).
func RandomInt(max int) int {
	return r.Intn(max)
}

// RandomBool generates a random boolean value.
// #nosec G404
func RandomBool() bool {
	n := 2 // #gomnd (golangci-lint magic number suppression)
	return r.Intn(n) == 1
}
