package utils

import "math/rand"

func generatePort() int {
	return rand.Intn(48128) + 1024
}
