package logger

import (
	"log"
	"os"
)

var (
	Std = log.New(os.Stdout, "", log.LstdFlags)
	Err = log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile)
)

func ResetLoggersFlags() {
	Std.SetFlags(0)
	Err.SetFlags(0)
}
