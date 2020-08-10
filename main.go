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

const version = "0.2.0"

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Режим отладки - детализирует этапы работы сервиса, также в данном режиме все логи отправляются в stdout",
			},
			&cli.StringFlag{
				Name:  "bind-address",
				Usage: "IP и порт слушаетля syslog, например: 0.0.0.0:514",
			},
			&cli.StringFlag{
				Name:  "log",
				Usage: "Файл, куда будут складываться логи",
				Value: constants.DEFAULT_LOG_FILE,
			},
			&cli.StringFlag{
				Name:  "maxmind",
				Usage: "Файл базы данных MaxMind с расширением .mmdb",
				Value: constants.DEFAULT_MAXMIND_DATABASE,
			},
			&cli.StringFlag{
				Name:  "influx-host",
				Usage: "URL подключения к Influx, например: http://0.0.0.0:8086",
			},
			&cli.StringFlag{
				Name:  "influx-db",
				Usage: "Название базы данных в Influx",
			},
		},
	}

	app.Action = func(c *cli.Context) error {
		var err error

		logger := lib.NewFileLogger(lib.LoggerConfig{
			Logfile: c.String("log"),
			IsDev:   c.Bool("debug"),
		})

		lib.StartupMessage(fmt.Sprintf("LimeHD Syslog Server v%s", version), logger)

		geoFinder, err := lib.NewGeoFinder(lib.GeoFinderConfig{
			MmdbPath: c.String("maxmind"),
			Logger:   logger,
		})

		if err != nil {
			logger.ErrorLog(err)
		}

		if len(c.String("bind-address")) == 0 {
			logger.ErrorLog(errors.New(constants.ADDRESS_IS_NOT_DEFINED))
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
		err = server.ListenUDP(c.String("bind-address"))

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
					if !logger.IsDevelopment() {
						logger.WarningLog(err)
						continue
					}

					logger.ErrorLog(err)
				}

				finderResult, err := geoFinder.Find(result.GetRemoteAddr())

				if err != nil {
					if !logger.IsDevelopment() {
						logger.WarningLog(err)
						continue
					}

					logger.ErrorLog(err)
				}

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
					if !logger.IsDevelopment() {
						logger.WarningLog(err)
						continue
					}

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
