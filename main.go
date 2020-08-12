package main

import (
	"fmt"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"github.com/LimeHD/limehd-syslog-server/lib"
	"github.com/urfave/cli"
	"gopkg.in/mcuadros/go-syslog.v2"
	"os"
)

const version = "0.3.3"

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Режим отладки - детализирует этапы работы сервиса, также в данном режиме все логи отправляются в stdout",
			},
			&cli.StringFlag{
				Name:     "bind-address",
				Usage:    "IP и порт слушаетля syslog, например: 0.0.0.0:514",
				Required: true,
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
				Name:     "influx-url",
				Usage:    "URL подключения к Influx, например: http://0.0.0.0:8086",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "influx-db",
				Usage:    "Название базы данных в Influx",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "influx-measurement",
				Usage:    "Название измерения (measurement) в Influx",
				Required: true,
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

		influx, err := lib.NewInfluxClient(lib.InfluxClientConfig{
			Addr:        c.String("influx-url"),
			Database:    c.String("influx-db"),
			Logger:      logger,
			Measurement: c.String("influx-measurement"),
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
						CountryName:  finderResult.GetCountryIsoCode(),
						Channel:      result.GetChannel(),
						StreamServer: result.GetStreamingServer(),
						Quality:      result.GetQuality(),
					},
					InfluxRequestFields: lib.InfluxRequestFields{
						BytesSent: result.GetBytesSent(),
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
