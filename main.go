package main

import (
	"errors"
	"fmt"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"github.com/LimeHD/limehd-syslog-server/lib"
	"github.com/urfave/cli"
	"gopkg.in/mcuadros/go-syslog.v2"
	"os"
)

const version = "0.1.0"

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "dev",
				Usage: "Is development mode run?",
			},
			&cli.StringFlag{
				Name:  "address",
				Usage: "host & port",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "log",
				Usage: "Log output file",
				Value: constants.DEFAULT_LOG_FILE,
			},
			&cli.StringFlag{
				Name:  "mmdb",
				Usage: "Maxmind mmdb database file",
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		var err error

		logger := lib.NewFileLogger(lib.LoggerConfig{
			Logfile: c.String("log"),
			IsDev:   c.Bool("dev"),
		})

		defer logger.Close()

		lib.StartupMessge(fmt.Sprintf("LimeHD Syslog Server v%s", version), logger)

		geoFinder, err := lib.NewGeoFinder(lib.GeoFinderConfig{
			MmdbPath: c.String("mmdb"),
			Logger:   logger,
		})

		if err != nil {
			logger.ErrorLog(err)
		}

		defer geoFinder.Close()

		if len(c.String("address")) == 0 {
			logger.ErrorLog(errors.New("Address is not defined"))
		}

		channel := make(syslog.LogPartsChannel)
		handler := syslog.NewChannelHandler(channel)

		server := syslog.NewServer()
		// RFC5424 - не подходит
		server.SetFormat(syslog.RFC3164)
		server.SetHandler(handler)
		err = server.ListenUDP(c.String("address"))

		if err != nil {
			logger.ErrorLog(err)
		}

		err = server.Boot()

		if err != nil {
			logger.ErrorLog(err)
		}

		parser := lib.NewSyslogParser(logger, lib.ParserConfig{
			PartsDelim:  constants.LOG_DELIM,
			StreamDelim: constants.REQUEST_URI_DELIM,
		})

		go func(channel syslog.LogPartsChannel) {
			for logParts := range channel {
				result, err := parser.Parse(logParts)

				if err != nil {
					logger.ErrorLog(err)
				}

				// пока для примера
				fmt.Println(result.GetBytesSent())
				fmt.Println(result.GetStreamingServer())
				fmt.Println(result.GetChannel())
				fmt.Println(result.GetQuality())
				fmt.Println(result.GetBytesSent())
				fmt.Println(result.GetRemoteAddr())

				finderResult, err := geoFinder.Find(result.GetRemoteAddr())

				if err != nil {
					logger.ErrorLog(err)
				}

				fmt.Println(finderResult.GetCountryGeoId())
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
