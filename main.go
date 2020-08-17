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
		Flags: CliFlags,
	}

	// todo вынести инициализацию всех компонентов
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
			ScheduleCallback: func(o *lib.Online) {
				channelConnections := o.Connections()
				err := influx.PointOnline(lib.InfluxOnlineRequestParams{
					Channels: channelConnections,
				})

				if err != nil {
					logger.ErrorLog(err)
				}

				if logger.IsDevelopment() {
					logger.InfoLog(fmt.Sprintf("Flushed connections: %d", channelConnections))
				}

				o.Flush()
			},
		})

		go online.Scheduler()
		go func(channel syslog.LogPartsChannel) {
			for logParts := range channel {
				result, err := parser.Parse(logParts)

				if err != nil {
					logger.ErrorLog(err)
					continue
				}

				finder, err := geoFinder.Find(result.GetRemoteAddr())

				if err != nil {
					logger.ErrorLog(err)
					continue
				}

				err = influx.Point(lib.InfluxRequestParams{
					InfluxRequestTags: lib.InfluxRequestTags{
						CountryName:  finder.GetCountryIsoCode(),
						AsnNumber:    finder.GetOrganizationNumber(),
						AsnOrg:       finder.GetOrganization(),
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
					logger.ErrorLog(err)
					continue
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
