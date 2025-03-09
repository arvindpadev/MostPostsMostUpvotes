package log

import (
	"io"
	"log"
)

var (
	Warn *log.Logger
	Info *log.Logger
	Error *log.Logger
	Debug *log.Logger
)

func InitLoggers(logLevel string) {
	errorWriter := log.Writer()
	warnWriter := log.Writer()
	infoWriter := log.Writer()
	debugWriter := io.Discard
	switch logLevel {
	case "error":
		warnWriter = io.Discard
		infoWriter = io.Discard
		break
	case "warn":
		infoWriter = io.Discard
		break
	case "debug":
		debugWriter = log.Writer()
		break
	}

	Error = log.New(errorWriter, "ERROR: ", log.Flags())
	Warn = log.New(warnWriter, "WARN: ", log.Flags())
	Info = log.New(infoWriter, "INFO: ", log.Flags())
	Debug = log.New(debugWriter, "DEBUG: ", log.Flags())
}
