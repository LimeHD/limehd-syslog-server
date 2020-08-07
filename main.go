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
				Usage: "development mode run",
			},
			&cli.StringFlag{
				Name:  "address",
				Usage: "host and ip address for connection syslog",
				Value: "",
			},
			&cli.StringFlag{
				Name:  "log",
				Usage: "file for log output",
				Value: constants.DEFAULT_LOG_FILE,
			},
			&cli.StringFlag{
				Name:  "maxmind",
				Usage: "MaxMind .mmdb database file",
				Value: "/usr/share/GeoIP/GeoLite2-City.mmdb",
			},
			&cli.StringFlag{
				Name:  "influx-host",
				Usage: "InfluxDB connection string",
			},
			&cli.StringFlag{
				Name:  "influx-db",
				Usage: "InfluxDB database name",
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		var err error

		logger := lib.NewFileLogger(lib.LoggerConfig{
			Logfile: c.String("log"),
			IsDev:   c.Bool("dev"),
		})

		lib.StartupMessge(fmt.Sprintf("LimeHD Syslog Server v%s", version), logger)

		geoFinder, err := lib.NewGeoFinder(lib.GeoFinderConfig{
			MmdbPath: c.String("maxmind"),
			Logger:   logger,
		})

		if err != nil {
			logger.ErrorLog(err)
		}

		if len(c.String("address")) == 0 {
			logger.ErrorLog(errors.New("Address is not defined"))
		}

		influx, err := lib.NewInfluxClient(lib.InfluxClientConfig{
			Addr:     c.String("influx-host"),
			Database: c.String("influx-db"),
			Logger:   logger,
		})

		if err != nil {
			logger.ErrorLog(err)
		}

		lib.Notifier(
			logger,
			geoFinder,
			influx,
		)

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
				fmt.Println(result.GetStreamingServer())
				fmt.Println(result.GetChannel())
				fmt.Println(result.GetQuality())
				fmt.Println(result.GetBytesSent())
				fmt.Println(result.GetRemoteAddr())
				fmt.Println(result.GetConnections())

				finderResult, err := geoFinder.Find("89.191.131.243")

				if err != nil {
					logger.ErrorLog(err)
				}

				fmt.Println(finderResult.GetCountryGeoId())

				err = influx.Point(lib.InfluxRequestParams{
					InfluxRequestTags: lib.InfluxRequestTags{
						CountryId:    finderResult.GetCountryGeoId(),
						Channel:      result.GetChannel(),
						StreamServer: result.GetStreamingServer(),
						Quality:      result.GetQuality(),
					},
					InfluxRequestFields: lib.InfluxRequestFields{
						BytesSent:   result.GetBytesSent(),
						Connections: result.GetConnections(),
					},
				})

				if err != nil {
					logger.ErrorLog(err)
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
