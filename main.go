package main

import "context"

func main() {
	ctx := context.Background()
	ExampleContextFunc(ctx)
}

func ExampleContextFunc(ctx context.Context) {}
