package main

import (
	"log"
	"os"
)

type Logger interface {
	ErrorLog(error error)
	Close()
}

type FileLogger struct {
	handler *os.File
	isDev bool
}

func NewFileLogger(logfile string, isDev bool) FileLogger {
	file, err := _createFileLogger(logfile)

	if err != nil {
		log.Fatal(err)
	}

	if !isDev {
		log.SetOutput(file)
	}

	return FileLogger{
		handler: file,
		isDev: isDev,
	}
}

func _createFileLogger(logfile string) (*os.File, error) {
	return os.OpenFile(logfile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
}

func (f FileLogger) ErrorLog(error error) {
	log.Fatalf("[LIMEHD SYSLOG ERROR]: %v", error)
}

func (f FileLogger) InfoLog(msg interface{})  {
	log.Printf("[LIMEHD SYSLOG INFO]: %v", msg)
}

func (f FileLogger) Close() {
	_ = f.handler.Close()
}

func (f FileLogger) IsDevelopment() bool  {
	return f.isDev
}
