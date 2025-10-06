package logger

import (
	"log"
)

func Info(message string) {
	log.Println("[INFO] " + message)
}

func Warning(message string) {
	log.Println("[WARN] " + message)
}

func Error(message string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %s", message, err.Error())
	} else {
		log.Printf("[ERROR] %s", message)
	}
}
