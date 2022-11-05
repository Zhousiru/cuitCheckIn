package cuit

import (
	"math/rand"
	"strconv"
	"time"
)

func genRandInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func genWinW() string {
	return strconv.Itoa(genRandInt(500, 1920))
}

func genWinH() string {
	return strconv.Itoa(genRandInt(500, 1080))
}
