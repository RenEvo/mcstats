package logging

import (
	"fmt"
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

func Debug(msg string, v ...interface{}) {
	logger.Printf("[DEBUG] %s", fmt.Sprintf(msg, v...))
}

func Error(msg string, v ...interface{}) {
	logger.Printf("[ERROR] %s", fmt.Sprintf(msg, v...))
}

func Panic(err error) {
	if err == nil {
		return
	}
	panic(err)
}
