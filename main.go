package main

import (
	"context"
	"fmt"
)

func main() {
	ctxParent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()

	ctxChild, cancelChild := context.WithCancel(ctxParent)
	defer cancelChild()

	// 親contextの中断は子contextにも伝播する
	cancelParent()
	// 子contextの中断は親contextには伝播しない
	// cancelChild()

	fmt.Printf("parent.Err is %v\n", ctxParent.Err())
	fmt.Printf("child.Err is %v\n", ctxChild.Err())
}
