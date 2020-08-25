package lib

import (
	"fmt"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"strconv"
	"time"
)

type (
	InfluxClient struct {
		c                 client.Client
		Database          string
		Measurement       string
		MeasurementOnline string
		_logger           Logger
	}

	InfluxClientConfig struct {
		Addr              string
		Database          string
		Logger            Logger
		Measurement       string
		MeasurementOnline string
	}

	InfluxRequestTags struct {
		CountryName  string
		AsnNumber    uint
		AsnOrg       string
		Channel      string
		StreamServer string
		Host         string
		Quality      string
		Time         time.Time
	}

	InfluxRequestFields struct {
		BytesSent   int
		Connections int
	}

	InfluxRequestParams struct {
		InfluxRequestTags
		InfluxRequestFields
	}

	InfluxOnlineRequestParams struct {
		Channels map[string]ChannelConnections
	}
)

func NewInfluxClient(config InfluxClientConfig) (*InfluxClient, error) {
	var err error
	i := &InfluxClient{}
	i.c, err = client.NewHTTPClient(client.HTTPConfig{
		Addr: config.Addr,
	})
	i.Database = config.Database
	i._logger = config.Logger
	i.Measurement = config.Measurement
	i.MeasurementOnline = config.MeasurementOnline

	if err != nil {
		return nil, err
	}

	d, s, err := i.c.Ping(time.Second * 5)

	if i._logger.IsDevelopment() {
		i._logger.InfoLog(fmt.Sprintf("Connect to Influx server version %s", s))
		i._logger.InfoLog(fmt.Sprintf("Connection duration is %v", d))
	}

	if err != nil {
		return nil, err
	}

	return i, nil
}

type tags map[string]string
type fields map[string]interface{}

func (i InfluxClient) Point(params []InfluxRequestParams) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: i.Database,
	})

	if err != nil {
		return err
	}

	for _, param := range params {
		pt, err := i.createPoint(i.Measurement,
			tags{
				"country_name":     param.CountryName,
				"asn_number":       strconv.FormatUint(uint64(param.AsnNumber), 10),
				"asn_org":          param.AsnOrg,
				"channel":          param.Channel,
				"streaming_server": param.StreamServer,
				"host":             param.Host,
				"quality":          param.Quality,
			},
			fields{
				"bytes_sent": param.BytesSent,
			},
			param.Time,
		)

		if err != nil {
			return err
		}

		bp.AddPoint(pt)
	}

	if err := i.c.Write(bp); err != nil {
		return err
	}

	return nil
}

func (i InfluxClient) PointOnline(params InfluxOnlineRequestParams) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: i.Database,
	})

	if err != nil {
		return err
	}

	// формируем данные пачками для отправки в influx
	for name, channel := range params.Channels {
		pt, err := i.createPoint(i.MeasurementOnline,
			tags{
				"channel": name,
			},
			fields{
				"value": channel.Count(),
			},
			time.Now(),
		)

		if err != nil {
			return err
		}

		bp.AddPoint(pt)
	}

	if err := i.c.Write(bp); err != nil {
		return err
	}

	return nil
}

// todo временную метку нужно брать с самого запроса
func (i InfluxClient) createPoint(m string, t tags, f fields, tt time.Time) (*client.Point, error) {
	return client.NewPoint(m, t, f, tt)
}

//

func (i InfluxClient) Close() {
	_ = i.c.Close()
}

func (i InfluxClient) CloseMessage() string {
	return "Close Influx connection"
}
