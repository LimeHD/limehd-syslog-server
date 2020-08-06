package main

import (
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/mcuadros/go-syslog.v2"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:   "dev",
				Usage:  "Is development mode run?",
			},
			&cli.StringFlag{
				Name:   "address",
				Usage:  "host & port",
				Value:  "",
			},
			&cli.StringFlag{
				Name:   "log",
				Usage:  "Log output file",
				Value:  "./tmp/log",
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		var err error
		fileLogger := NewFileLogger(c.String("log"), c.Bool("dev"))
		defer fileLogger.Close()

		fileLogger.InfoLog("LimeHD Syslog Server v0.1.0")

		if len(c.String("address")) == 0 {
			fileLogger.ErrorLog(errors.New("Address is not defined"))
		}

		channel := make(syslog.LogPartsChannel)
		handler := syslog.NewChannelHandler(channel)

		server := syslog.NewServer()
		// RFC5424 - не подходит
		server.SetFormat(syslog.RFC3164)
		server.SetHandler(handler)
		err = server.ListenUDP(c.String("address"))

		if err != nil {
			fileLogger.ErrorLog(err)
		}

		err = server.Boot()

		if err != nil {
			fileLogger.ErrorLog(err)
		}

		go func(channel syslog.LogPartsChannel) {
			for logParts := range channel {
				if fileLogger.IsDevelopment() {
					fmt.Println(logParts)
				}

			}
		}(channel)

		server.Wait()

		return err
	}

	err := app.Run(os.Args)

	if err != nil {
		// todo signal notify
		fmt.Println(err)
	}
}
