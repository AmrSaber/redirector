package logger

import (
	"log"
	"os"
)

var (
	Std = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	Err = log.New(os.Stderr, "error: ", log.LstdFlags|log.Lshortfile)
)
