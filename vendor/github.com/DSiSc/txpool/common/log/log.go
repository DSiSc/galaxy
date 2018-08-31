package log

import (
	"log"
)

func Info(format string) {
	log.Printf("[Info] %s\n", format)
}

func Warn(format string) {
	log.Printf("[Warning] %s\n", format)
}

func Error(format string) {
	log.Printf("[Error] %s\n", format)
}
