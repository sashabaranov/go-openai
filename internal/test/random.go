package test

import (
	"math/rand"
	"os"
	"strconv"
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

// MaybeSeedRNG optionally seeds the random number generator based on the
// TEST_RNG_SEED environment variable. If TEST_RNG_SEED is set to an integer
// value, the RNG is seeded with that value. If it's set to "random", the RNG
// is seeded using the current time. If the variable is not set or contains an
// invalid value, the RNG remains unseeded and retains its default behavior.
func MaybeSeedRNG() {
	seedEnv := os.Getenv("TEST_RNG_SEED")

	if seedValue, err := strconv.ParseInt(seedEnv, 10, 64); err == nil {
		rand.Seed(seedValue)
		return
	}

	if seedEnv == "random" {
		rand.Seed(time.Now().UnixNano())
	}
}
