package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandomInt returns a random integer between min and max (inclusive)
func RandomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

// Shuffle shuffles a slice of any type (using generic approach if Go version supports, else interface)
// For simplicity and compatibility with older Go in this environment:
func Shuffle(slice interface{}) {
	switch s := slice.(type) {
	case []uint:
		rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
	case []int:
		rand.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
	}
}

// Note: Shuffle for users specifically if needed, or use a more generic approach
