package logger

import (
	"log"
	"os"
)

func NewInfoLogger() *log.Logger {
	return log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
}

func NewErrorLogger() *log.Logger {
	return log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
