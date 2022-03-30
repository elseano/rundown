package util

import (
	"encoding/hex"
	"math/rand"
)

func RandomString() string {
	result := make([]byte, 6)
	rand.Read(result)

	return hex.EncodeToString(result)
}
