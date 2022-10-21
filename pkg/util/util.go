package util

import (
	"context"
	"math/rand"
	"time"
)

const jitter = 50

// 设置随机延迟，防止并发探测量过大
func SetJitter() int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(jitter)
}

func Wait(ctx context.Context, interval time.Duration, f func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	f()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f()
		}
	}
}
