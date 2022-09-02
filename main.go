package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	wg := &sync.WaitGroup{}

	wg.Add(2)

	go func() {
		defer wg.Done()

		for i := 0; i < 3; i++ {
			fmt.Printf("wg 1: %d / 3\n", i+1)
			time.Sleep(1 * time.Second)
		}
	}()

	go func() {
		defer wg.Done()

		for i := 0; i < 5; i++ {
			fmt.Printf("wg 2: %d / 5\n", i+1)
			time.Sleep(1 * time.Second)
		}
	}()

	wg.Wait()
	fmt.Println("wg: done")
}
