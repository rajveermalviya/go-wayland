package main

import (
	"log"
	"os"
)

var logDisabled = os.Getenv("LOG_DISABLED") == "1"

func logPrintln(v ...interface{}) {
	if !logDisabled {
		log.Println(v...)
	}
}

func logPrintf(format string, v ...interface{}) {
	if !logDisabled {
		log.Printf(format, v...)
	}
}
