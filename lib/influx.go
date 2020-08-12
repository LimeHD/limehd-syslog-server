package lib

import (
	"fmt"
	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"time"
)

type (
	InfluxClient struct {
		c           client.Client
		Database    string
		Measurement string
		_logger     Logger
	}

	InfluxClientConfig struct {
		Addr        string
		Database    string
		Logger      Logger
		Measurement string
	}

	InfluxRequestTags struct {
		CountryName  string
		Channel      string
		StreamServer string
		Quality      string
	}

	InfluxRequestFields struct {
		BytesSent   int
		Connections int
	}

	InfluxRequestParams struct {
		InfluxRequestTags
		InfluxRequestFields
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

func (i InfluxClient) Point(params InfluxRequestParams) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database: i.Database,
	})

	if err != nil {
		return err
	}

	pt, err := i.CreatePoint(i.Measurement,
		tags{
			"country_name":     params.CountryName,
			"channel":          params.Channel,
			"streaming_server": params.StreamServer,
			"quality":          params.Quality,
		},
		fields{
			"bytes_sent": params.BytesSent,
		},
	)

	bp.AddPoint(pt)

	if err := i.c.Write(bp); err != nil {
		return err
	}

	return nil
}

func (i InfluxClient) CreatePoint(m string, t tags, f fields) (*client.Point, error) {
	return client.NewPoint(m, t, f, time.Now())
}

//

func (i InfluxClient) Close() {
	_ = i.c.Close()
}

func (i InfluxClient) CloseMessage() string {
	return "Close Influx connection"
}
