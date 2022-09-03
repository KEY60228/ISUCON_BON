package main

import (
	"log"
	"os"
)

var (
	ContestantLogger = log.New(os.Stdout, "", log.Ltime|log.Lmicroseconds)
	AdminLogger      = log.New(os.Stderr, "[ADMIN]", log.Ltime|log.Lmicroseconds)
)
