package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

L:
	for i := 0; ; i++ {
		fmt.Printf("loop %d\n", i)

		select {
		case <-ctx.Done():
			break L
		case <-time.After(1 * time.Second):
		}
	}
}
