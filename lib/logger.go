package lib

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	ErrorLog(error error)
	InfoLog(msg interface{})
	IsDevelopment() bool
	Close()
}

type LoggerConfig struct {
	Logfile string
	IsDev   bool
}

type FileLogger struct {
	handler *os.File
	isDev   bool
}

func NewFileLogger(config LoggerConfig) Logger {
	file, err := _createFileLogger(config.Logfile)

	if err != nil {
		log.Fatal(err)
	}

	if !config.IsDev {
		log.SetOutput(file)
	}

	return FileLogger{
		handler: file,
		isDev:   config.IsDev,
	}
}

func _createFileLogger(logfile string) (*os.File, error) {
	return os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}

func (f FileLogger) ErrorLog(error error) {
	log.Fatalf("[LIMEHD SYSLOG ERROR]: %v", error)
}

func (f FileLogger) InfoLog(msg interface{}) {
	log.Printf("[LIMEHD SYSLOG INFO]: %v", msg)
}

func (f FileLogger) Close() {
	_ = f.handler.Close()
}

func (f FileLogger) IsDevelopment() bool {
	return f.isDev
}

func (f FileLogger) CloseMessage() string {
	return "Close File logger"
}

func StartupMessge(message string, logger Logger) {
	logger.InfoLog(message)
	// заодно выведем и на stdout

	if !logger.IsDevelopment() {
		fmt.Println(message)
	}
}