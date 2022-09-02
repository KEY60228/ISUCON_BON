package main

import (
	"fmt"
	"sync"
)

func main() {
	userIDs := make([]int, 0)
	userIDsLock := &sync.Mutex{}

	wg := &sync.WaitGroup{}

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			userIDsLock.Lock()
			userIDs = append(userIDs, id)
			userIDsLock.Unlock()
		}(i)
	}

	wg.Wait()

	fmt.Printf("userIDs: %v\n", userIDs)
}
