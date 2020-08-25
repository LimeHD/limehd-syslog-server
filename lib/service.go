package lib

import (
	"github.com/LimeHD/limehd-syslog-server/constants"
	"github.com/urfave/cli"
)

type Service struct {
	logger   Logger
	influx   *InfluxClient
	finder   *GeoFinder
	parser   *SyslogParser
	stream   *StreamQueue
	template *Template
	// todo
	// online, pool
}

func NewService(c *cli.Context) *Service {
	var err error

	s := new(Service)

	s.logger = NewFileLogger(
		LoggerConfig{
			Logfile: c.String("log"),
			IsDev:   c.Bool("debug"),
		},
	)

	geoFinder, err := NewGeoFinder(
		GeoFinderConfig{
			MmdbPath:    c.String("maxmind"),
			AsnMmdbPath: c.String("maxmind-asn"),
			Logger:      s.logger,
		},
	)

	if err != nil {
		s.logger.ErrorLog(err)
	}

	influx, err := NewInfluxClient(
		InfluxClientConfig{
			Addr:              c.String("influx-url"),
			Database:          c.String("influx-db"),
			Logger:            s.logger,
			Measurement:       c.String("influx-measurement"),
			MeasurementOnline: c.String("influx-measurement-online"),
		},
	)

	if err != nil {
		s.logger.ErrorLog(err)
	}

	template, err := NewTemplate(
		TemplateConfig{
			Template: c.String("nginx-template"),
		},
	)

	if err != nil {
		s.logger.ErrorLog(err)
	}

	parser := NewSyslogParser(
		s.logger,
		ParserConfig{
			PartsDelim:  constants.LOG_DELIM,
			StreamDelim: constants.REQUEST_URI_DELIM,
			Template:    template,
		},
	)

	stream := NewStream()

	Notifier(
		s.logger,
		geoFinder,
		influx,
	)

	s.influx = influx
	s.finder = &geoFinder
	s.parser = &parser
	s.stream = stream

	return s
}

func (s Service) GetLogger() Logger {
	return s.logger
}

func (s Service) GetInfluxClient() *InfluxClient {
	return s.influx
}

func (s Service) GetParser() *SyslogParser {
	return s.parser
}

func (s Service) GetFinder() *GeoFinder {
	return s.finder
}

func (s Service) GetStream() *StreamQueue {
	return s.stream
}
