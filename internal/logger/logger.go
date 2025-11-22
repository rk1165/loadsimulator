package logger

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/rk1165/loadsimulator/internal/load"
)

var (
	InfoLog  *log.Logger
	WarnLog  *log.Logger
	ErrorLog *log.Logger
)

func init() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)

	InfoLog = log.New(multiWriter, "   INFO: ", log.Ldate|log.Ltime)
	WarnLog = log.New(multiWriter, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(multiWriter, "  ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func CreateLoadLog(file string) load.Log {
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("failed to create logs directory: %v", err)
	}
	logFile, err := os.OpenFile(fmt.Sprintf("%s/%s.log", logsDir, file), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("failed to create log file %s error: %v", file, err)
	}
	writer := io.Writer(logFile)
	InfoLog = log.New(writer, "   INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(writer, "  ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	return load.Log{InfoLog: InfoLog, ErrorLog: ErrorLog}
}
