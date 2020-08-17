package lib

import (
	"fmt"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"log"
	"os"
)

type Logger interface {
	ErrorLog(error error)
	InfoLog(msg interface{})
	WarningLog(msg interface{})
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
	handler := os.Stdout

	if len(config.Logfile) > 0 {
		if !fileExists(config.Logfile) {
			log.Println(constants.LOG_FILE_ON_EXIST)
		}

		handler, err := createFileLogger(config.Logfile)

		if err != nil {
			log.Fatal(err)
		}

		if !config.IsDev {
			log.SetOutput(handler)
		}
	}

	return FileLogger{
		handler: handler,
		isDev:   config.IsDev,
	}
}

func createFileLogger(logfile string) (*os.File, error) {
	return os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
}

func (f FileLogger) ErrorLog(error error) {
	if f.IsDevelopment() {
		log.Fatalf("[LIMEHD SYSLOG ERROR]: %v", error)
	} else {
		f.WarningLog(error)
	}
}

func (f FileLogger) InfoLog(msg interface{}) {
	log.Printf("[LIMEHD SYSLOG INFO]: %v", msg)
}

func (f FileLogger) WarningLog(msg interface{}) {
	log.Printf("[LIMEHD SYSLOG WARNING]: %v", msg)
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

func StartupMessage(message string, logger Logger) {
	logger.InfoLog(message)
	// заодно выведем и на stdout

	if !logger.IsDevelopment() {
		fmt.Println(message)
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
