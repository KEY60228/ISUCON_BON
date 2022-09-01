package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctxMain := context.Background()

	go func() {
		ctxTimeout, cancelTimeout := context.WithTimeout(ctxMain, 5*time.Second)
		defer cancelTimeout()
		<-ctxTimeout.Done()
		fmt.Println("timeout!")
	}()

	go func() {
		ctxDeadline, cancelDeadline := context.WithDeadline(ctxMain, time.Now().Add(3*time.Second))
		defer cancelDeadline()
		<-ctxDeadline.Done()
		fmt.Println("deadline!")
	}()

	for i := 0; i < 10; i++ {
		fmt.Printf("%d sec...\n", i)
		time.Sleep(1 * time.Second)
	}
}
