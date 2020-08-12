package main

import (
	"fmt"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"github.com/LimeHD/limehd-syslog-server/lib"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

// TODO generate full nginx formatted string
var nginxLogFormatSlice = []string{
	"11/Aug/2020:14:01:32 +0300", "1597143692.596",
	// request
	"127.0.0.1", "HTTP/1.1", "GET", "syslog-server.local", "/streaming/domashniy/324/vh1w/playlist.m3u8", "-",
	// response
	"404", "209",
	// request time
	"0.000",
	// upstream
	"-", "-", "-",
	// http
	"-", "-", "-", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36", "-",
	// connection
	"3", "84",
	// bytes sent
	"404",
}

var clients = []string{
	"127.0.0.1:38001",
	"0.0.0.0:38001",
	"::1",
	"",
}

var tags = []string{
	"nginx", "",
}

func _getNginxFormatString() string {
	return strings.Join(nginxLogFormatSlice, "|")
}

func _getRandomString(s []string) string {
	randomIndex := rand.Intn(len(s))
	return s[randomIndex]
}

func _generateRandomParts() map[string]interface{} {
	randomParts := map[string]interface{}{
		"client":    _getRandomString(clients),
		"content":   _getNginxFormatString(),
		"facility":  0,
		"hostname":  "ppc",
		"priority":  0,
		"severity":  0,
		"tag":       _getRandomString(tags),
		"timestamp": time.Now().String(),
		"tls_peer":  "",
	}

	return randomParts
}

func generate(parser lib.SyslogParser, logger lib.Logger, logs chan<- lib.Log, wg *sync.WaitGroup) {
	parts := _generateRandomParts()
	result, err := parser.Parse(parts)

	if err != nil {
		logger.ErrorLog(err)
	}

	logs <- result

	wg.Done()
}

func see(logs <-chan lib.Log) {
	for log := range logs {
		fmt.Println(log)
	}
}

func TestLog(t *testing.T) {
	logger := lib.NewFileLogger(lib.LoggerConfig{
		Logfile: "./tmp/log",
		IsDev:   false,
	})
	parser := lib.NewSyslogParser(logger, lib.ParserConfig{
		PartsDelim:  constants.LOG_DELIM,
		StreamDelim: constants.REQUEST_URI_DELIM,
	})

	wg := &sync.WaitGroup{}
	logs := make(chan lib.Log, 11)

	for i := 0; i <= 10; i++ {
		wg.Add(1)
		go generate(parser, logger, logs, wg)
	}

	go see(logs)

	wg.Wait()
}
