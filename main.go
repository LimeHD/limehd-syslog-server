package main

import (
	"fmt"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"github.com/LimeHD/limehd-syslog-server/lib"
	"github.com/urfave/cli"
	"gopkg.in/mcuadros/go-syslog.v2"
	"os"
)

const version = "0.3.11" // Automaticaly updated // Automaticaly updated // Automaticaly updated // Automaticaly updated

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
				Name:     "log",
				Usage:    "Файл, куда будут складываться логи",
				Required: false,
			},
			&cli.StringFlag{
				Name:  "maxmind",
				Usage: "Файл базы данных MaxMind с расширением .mmdb",
				Value: constants.DEFAULT_MAXMIND_DATABASE,
			},
			&cli.StringFlag{
				Name:  "maxmind-asn",
				Usage: "Файл базы данных MaxMind ASN с расширением .mmdb для автономных систем",
				Value: constants.DEFAULT_MAXMIND_ASN_DATABASE,
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
			&cli.StringFlag{
				Name:     "influx-measurement-online",
				Usage:    "Название измерения (measurement) в Influx для счетчиков online пользователей",
				Required: true,
			},
			&cli.Int64Flag{
				Name:     "online-duration",
				Usage:    "За какой промежуток агрегировать уникальных пользователей (в секундах)",
				Value:    300,
				Required: true,
			},
			&cli.StringFlag{
				Name:  "nginx-template",
				Usage: "Шаблон для конфигурации форматов логов Nginx",
				Value: "./template.conf",
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
			MmdbPath:    c.String("maxmind"),
			AsnMmdbPath: c.String("maxmind-asn"),
			Logger:      logger,
		})

		if err != nil {
			logger.ErrorLog(err)
		}

		influx, err := lib.NewInfluxClient(lib.InfluxClientConfig{
			Addr:              c.String("influx-url"),
			Database:          c.String("influx-db"),
			Logger:            logger,
			Measurement:       c.String("influx-measurement"),
			MeasurementOnline: c.String("influx-measurement-online"),
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

		template, err := lib.NewTemplate(lib.TemplateConfig{
			Template: c.String("nginx-template"),
		})

		if err != nil {
			logger.ErrorLog(err)
		}

		parser := lib.NewSyslogParser(logger, lib.ParserConfig{
			PartsDelim:  constants.LOG_DELIM,
			StreamDelim: constants.REQUEST_URI_DELIM,
			Template:    template,
		})

		online := lib.NewOnline(lib.OnlineConfig{
			OnlineDuration: c.Int64("online-duration"),
		})

		go online.Scheduler(func() {
			channelConnections := online.Connections()
			err := influx.PointOnline(lib.InfluxOnlineRequestParams{
				Channels: channelConnections,
			})

			if err != nil {
				if !logger.IsDevelopment() {
					logger.WarningLog(err)
				} else {
					logger.ErrorLog(err)
				}
			}

			if logger.IsDevelopment() {
				logger.InfoLog(fmt.Sprintf("Flushed connections: %d", channelConnections))
			}

			online.Flush()
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
						AsnNumber:    finderResult.GetOrganizationNumber(),
						AsnOrg:       finderResult.GetOrganization(),
						Channel:      result.GetChannel(),
						StreamServer: result.GetClientAddr(),
						Host:         result.GetStreamingServer(),
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

				// Пользователи онлайн

				unique := lib.UniqueIdentity{
					Channel: result.GetChannel(),
					UniqueCombination: lib.UniqueCombination{
						Ip:        result.GetRemoteAddr(),
						UserAgent: result.GetUserAgent(),
					},
				}

				online.Peek(unique)

				// todo удалить
				if logger.IsDevelopment() {
					logger.InfoLog(online.Connections())
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
