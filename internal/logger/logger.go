package logger

import (
	"log"
)

func Info(msg string, args ...any) {
	log.Printf("[INFO] "+msg, args...)
}

func Error(msg string, args ...any) {
	log.Printf("[ERROR] "+msg, args...)
}

func Worker(id int, msg string, args ...any) {
	log.Printf("[WORKER-%d] "+msg, append([]any{id}, args...)...)
}
