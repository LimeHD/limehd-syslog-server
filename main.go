package main

import (
	"fmt"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"github.com/LimeHD/limehd-syslog-server/lib"
	"github.com/urfave/cli"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
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
					logger.InfoLog(fmt.Sprintf("Flushed connections: %v", channelConnections))
				}

				logger.InfoLog(fmt.Sprintf("Total %d", o.Total()))
				logger.InfoLog(o)

				o.Flush()
			},
		})

		sendToInfluxCallback := func(receive lib.Receiver) {
			err = influx.Point(lib.InfluxRequestParams{
				InfluxRequestTags: lib.InfluxRequestTags{
					CountryName:  receive.Finder.GetCountryIsoCode(),
					AsnNumber:    receive.Finder.GetOrganizationNumber(),
					AsnOrg:       receive.Finder.GetOrganization(),
					Channel:      receive.Parser.GetChannel(),
					StreamServer: receive.Parser.GetClientAddr(),
					Host:         receive.Parser.GetStreamingServer(),
					Quality:      receive.Parser.GetQuality(),
				},
				InfluxRequestFields: lib.InfluxRequestFields{
					BytesSent: receive.Parser.GetBytesSent(),
				},
			})

			if err != nil {
				logger.ErrorLog(err)
				return
			}

			// Пользователи онлайн
			unique := lib.UniqueIdentity{
				Channel: receive.Parser.GetChannel(),
				UniqueCombination: lib.UniqueCombination{
					Ip:        receive.Parser.GetRemoteAddr(),
					UserAgent: receive.Parser.GetUserAgent(),
				},
			}

			online.Peek(unique)

			// todo удалить
			if logger.IsDevelopment() {
				logger.InfoLog(online.Connections())
			}
		}

		receiveAndParseLogsCallback := func(p format.LogParts) (lib.Receiver, error) {
			result, err := parser.Parse(p)

			if err != nil {
				logger.ErrorLog(err)
				return lib.Receiver{}, err
			}

			finder, err := geoFinder.Find(result.GetRemoteAddr())

			if err != nil {
				logger.ErrorLog(err)
				return lib.Receiver{}, err
			}

			return lib.Receiver{
				Parser: result,
				Finder: finder,
			}, nil
		}

		pool := lib.NewPool(lib.PoolConfig{
			ListenerCallback: sendToInfluxCallback,
			ReceiverCallback: receiveAndParseLogsCallback,
			PoolSize:         c.Int("pool-size"),
			WorkersCount:     c.Int("worker-count"),
			SenderCount:      c.Int("sender-count"),
			WorkerPoolSize:   c.Int("worker-pool-size"),
			WorkerFn: func(p *lib.Pool, channel syslog.LogPartsChannel) {
				for logParts := range channel {
					p.Task(logParts)
				}
			},
		})

		go online.Scheduler()
		go func(channel syslog.LogPartsChannel) {
			pool.Run(channel, c.Int("max-parallel"))
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
