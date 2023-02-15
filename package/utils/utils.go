package utils

import (
	"fmt"
	"math/rand"
	"time"
)

func generatePort() int {
	return rand.Intn(48128) + 1024
}

func timeStr() string {
	currentTime := time.Now()
	return fmt.Sprintf("%02d:%02d:%02d.%03d",
		currentTime.Hour(),
		currentTime.Hour(),
		currentTime.Second(),
		currentTime.Nanosecond()/1000_000)
}
