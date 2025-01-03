package logs

import (
	"fmt"
	"log"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	SuccessLog *log.Logger
	ErrorLog   *log.Logger
)

func Logger() {
	successLogger := &lumberjack.Logger{
		Filename:   "./logs/success.log",
		MaxSize:    5,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	errorLogger := &lumberjack.Logger{
		Filename:   "./logs/error.log",
		MaxSize:    5,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   true,
	}

	SuccessLog = log.New(successLogger, "SUCCESS: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(errorLogger, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	fmt.Println("Logger initialized")
}
