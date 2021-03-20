package log

import (
	"fmt"
	"log"
	"os"
)

var (
	debug  = false
	logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Llongfile)
)

func init() {
	debug = os.Getenv("WAYLAND_DEBUG") == "1"
}

func Print(v ...interface{}) {
	if debug {
		_ = logger.Output(2, fmt.Sprint(v...))
	}
}

func Printf(format string, v ...interface{}) {
	if debug {
		_ = logger.Output(2, fmt.Sprintf(format, v...))
	}
}

func Println(v ...interface{}) {
	if debug {
		_ = logger.Output(2, fmt.Sprintln(v...))
	}
}

func Fatal(v ...interface{}) {
	_ = logger.Output(2, fmt.Sprint(v...))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	_ = logger.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func Fatalln(v ...interface{}) {
	_ = logger.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}
