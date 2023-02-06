package util

import (
	"fmt"
	"math/rand"
	"time"
)

func RandInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano() * 34751397)
	return min + rand.Intn(max-min+1)
}

func GetOTP(digits uint) string {
	var otp int
	var min int = 1
	var max int = 9

	for i := 1; i < int(digits); i++ {
		min = min * 10
		max = (max * 10) + 9
	}

	otp = RandInt(min, max)
	return fmt.Sprintf("%v", otp)
}
