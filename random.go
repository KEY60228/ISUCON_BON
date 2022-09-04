package main

import "math/rand"

var (
	randomStringPrefixes = []string{
		"Hello",
		"Hi",
		"Yay",
		"Oh",
		"Wow",
	}
	randomStringSuffixes = []string{
		"World",
		"Baby",
		"Image",
		"My Photo",
		"Great Picture",
	}
)

func randomText() string {
	prefix := randomStringPrefixes[rand.Intn(len(randomStringPrefixes))]
	suffix := randomStringSuffixes[rand.Intn(len(randomStringSuffixes))]
	return prefix + ", " + suffix
}
