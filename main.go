package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go LoopWithBefore(ctx)
	go LoopWithAfter(ctx)

	<-ctx.Done()
}

func LoopWithBefore(ctx context.Context) {
	beforeLoop := time.Now()
	for {
		loopTimer := time.After(3 * time.Second)

		HeavyProcess(ctx, "BEFORE")

		select {
		case <-ctx.Done():
			return
		case <-loopTimer:
			fmt.Printf("[Before] loop duration: %.2fs\n", time.Now().Sub(beforeLoop).Seconds())
			beforeLoop = time.Now()
		}
	}
}

func LoopWithAfter(ctx context.Context) {
	beforeLoop := time.Now()
	for {
		HeavyProcess(ctx, "AFTER")

		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
			fmt.Printf("[AFTER] loop duration: %.2fs\n", time.Now().Sub(beforeLoop).Seconds())
			beforeLoop = time.Now()
		}
	}
}

func HeavyProcess(ctx context.Context, pattern string) {
	fmt.Printf("[%s] Heavy Process\n", pattern)
	time.Sleep(1*time.Second + 500*time.Millisecond)
}
