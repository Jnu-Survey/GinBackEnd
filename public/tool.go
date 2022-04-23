package public

import (
	"math/rand"
	"time"
)

// RandomInt 生成随机数
func RandomInt(want int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(want)
}
