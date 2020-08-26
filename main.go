package main

import (
	"errors"
	"fmt"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"github.com/LimeHD/limehd-syslog-server/lib"
	"github.com/urfave/cli"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
	"os"
	"time"
)

const version = "0.3.11" // Automaticaly updated // Automaticaly updated // Automaticaly updated // Automaticaly updated

func main() {
	app := &cli.App{
		Flags: CliFlags,
	}

	// todo вынести инициализацию всех компонентов
	app.Action = func(c *cli.Context) error {
		var err error

		service := lib.NewService(c)
		logger := service.GetLogger()
		finder := service.GetFinder()
		influx := service.GetInfluxClient()
		parser := service.GetParser()
		stream := service.GetStream()
		online := service.GetOnline()

		lib.StartupMessage(fmt.Sprintf("LimeHD Syslog Server v%s", version), logger)

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

		aggregationCallback := func(receive lib.Receiver) error {
			// трафик
			stream.Add(lib.InfluxRequestParams{
				InfluxRequestTags: lib.InfluxRequestTags{
					CountryName:  receive.Finder.GetCountryIsoCode(),
					AsnNumber:    receive.Finder.GetOrganizationNumber(),
					AsnOrg:       receive.Finder.GetOrganization(),
					Channel:      receive.Parser.GetChannel(),
					StreamServer: receive.Parser.GetClientAddr(),
					Host:         receive.Parser.GetStreamingServer(),
					Quality:      receive.Parser.GetQuality(),
					Time:         time.Now(),
				},
				InfluxRequestFields: lib.InfluxRequestFields{
					BytesSent: receive.Parser.GetBytesSent(),
				},
			})

			// Пользователи онлайн
			unique := lib.UniqueIdentity{
				Channel: receive.Parser.GetChannel(),
				UniqueCombination: lib.UniqueCombination{
					Ip:        receive.Parser.GetRemoteAddr(),
					UserAgent: receive.Parser.GetUserAgent(),
				},
			}

			online.Peek(unique)
			logger.Debug(online.Connections())

			return nil
		}

		receiveAndParseLogsCallback := func(p format.LogParts) (lib.Receiver, error) {
			result, err := parser.Parse(p)

			if err != nil {
				return lib.Receiver{}, err
			}

			if !result.IsAvailableUri() {
				return lib.Receiver{}, errors.New(fmt.Sprintf("%s: %s", constants.NOT_AVAILABLE_URI, result.GetUri()))
			}

			finderResult, err := finder.Find(result.GetRemoteAddr())

			if err != nil {
				return lib.Receiver{}, err
			}

			return lib.Receiver{
				Parser: result,
				Finder: finderResult,
			}, nil
		}

		pool := lib.NewPool(
			lib.PoolConfig{
				ListenerCallback: aggregationCallback,
				ReceiverCallback: receiveAndParseLogsCallback,
				PoolSize:         c.Int("pool-size"),
				WorkersCount:     c.Int("worker-count"),
				SenderCount:      c.Int("sender-count"),
				WorkerPoolSize:   c.Int("worker-pool-size"),
				ErrorPoolSize:    c.Int("error-pool-size"),
				WorkerFn: func(p *lib.Pool, channel syslog.LogPartsChannel) {
					for logParts := range channel {
						p.Task(logParts)
					}
				},
				ErrorHandleCallback: func(err error) {
					logger.ErrorLog(err)
				},
				ErrorHandlerCount: c.Int("error-handler-count"),
			},
		)

		online.SetScheduleHandler(func(o *lib.Online) {
			// запрашиваем агрегацию
			channelConnections := o.Connections()
			// передаем управление
			o.Flush()

			err := influx.PointOnline(lib.InfluxOnlineRequestParams{
				Channels: channelConnections,
			})

			if err != nil {
				logger.ErrorLog(err)
			}

			logger.Debug(fmt.Sprintf("Flushed connections: %v", channelConnections))
			logger.InfoLog(fmt.Sprintf("Total %d", o.Total()))
			logger.InfoLog(o)
		})

		stream.SetScheduleHandler(func(s *lib.StreamQueue) {
			// аолучаем накопленные данные
			streams := s.All()
			// отдаем управление для нового накопления
			s.Flush()

			if err := influx.Point(streams); err != nil {
				logger.ErrorLog(err)
			}

			logger.InfoLog("Stream scheduler done!")
		})

		go online.Scheduler(c.Int64("online-duration"))
		go stream.Scheduler(c.Int("stream-duration"))

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
