package test

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const strLen = 10
const bitLen = 0xFF

// See StackOverflow answer:
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
// RandomString generates a cryptographically secure random string of length
// strLen.
func RandomString() string {
	sb := strings.Builder{}
	sb.Grow(strLen)

	for i := 0; i < strLen; i++ {
		randomByte := make([]byte, 1)
		// #nosec G404
		_, err := rand.Read(randomByte)
		if err != nil {
			log.Fatalf("Error generating random string: %v", err)
		}
		randomIndex := randomByte[0] % byte(len(letters))
		sb.WriteByte(letters[randomIndex])
	}

	return sb.String()
}

// RandomInt generates a random integer between 0 (inclusive) and 'max'
// (exclusive). We uses the crypto/rand library for generating random
// bytes. It then performs a bitwise AND operation with 0xFF to keep only the
// least significant 8 bits, effectively converting the byte to an integer. The
// resulting integer is then modulo'd with 'max'.
func RandomInt(max int) int {
	var b [1]byte
	// #nosec G404
	_, err := rand.Read(b[:])
	if err != nil {
		log.Fatalf("Error generating random int: %v", err)
	}
	n := int(b[0]&bitLen) % max
	return n
}

// RandomBool generates a cryptographically secure random boolean value.
// It reads a single byte from the crypto/rand library and uses its least
// significant bit to determine the boolean value. The function returns
// true if the least significant bit is 1, and false otherwise.
func RandomBool() bool {
	var b [1]byte
	// #nosec G404
	_, err := rand.Read(b[:])
	if err != nil {
		log.Fatalf("Error generating random bool: %v", err)
	}
	return b[0]&1 == 1
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
