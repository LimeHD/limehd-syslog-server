package lib

import (
	"github.com/DataDog/datadog-go/statsd"
	"strconv"
)

type (
	Statsd struct {
		client            *statsd.Client
		measurement       string
		measurementOnline string
		logger            Logger
	}
	StatsdConfig struct {
		Address           string
		Measurement       string
		MeasurementOnline string
		Logger            Logger
	}
	StatsdStreamingTags struct {
		CountryName  string
		AsnNumber    uint
		AsnOrg       string
		Channel      string
		StreamServer string
		Host         string
		Quality      string
	}
	StatsdOnlineTags struct {
		Channels map[string]ChannelConnections
	}
	Tag struct {
		tags []string
	}
)

func NewStatsdClient(c StatsdConfig) (Statsd, error) {
	s := Statsd{}

	ssdclient, err := statsd.New(c.Address,
		statsd.WithTags([]string{"env:dev", "service:limehd-syslog-server"}),
	)
	if err != nil {
		return s, err
	}

	s.client = ssdclient
	s.measurement = c.Measurement
	s.measurementOnline = c.MeasurementOnline

	return s, nil
}

func (s Statsd) Point(v float64, tags StatsdStreamingTags) error {
	t := newTags()
	t.append("country_name", tags.CountryName)
	t.append("asn_number", strconv.FormatUint(uint64(tags.AsnNumber), 10))
	t.append("asn_org", tags.AsnOrg)
	t.append("channel", tags.Channel)
	t.append("streaming_server", tags.StreamServer)
	t.append("host", tags.Host)
	t.append("quality", tags.Quality)

	return s.gause(s.measurement, v, t.tags, 1)
}

func (s Statsd) OnlinePoint(tags StatsdOnlineTags) {
	for name, channel := range tags.Channels {
		t := newTags()
		t.append("channel", name)

		if err := s.gause(s.measurementOnline, float64(channel.Count()), t.tags, 1); err != nil {
			s.logger.ErrorLog(err)
		}
	}
}

func (s Statsd) gause(name string, value float64, tags []string, rate float64) error {
	return s.client.Gauge(name, value, tags, rate)
}

func (s Statsd) Close() {
	_ = s.client.Close()
}

func (s Statsd) CloseMessage() string {
	return "Close statsd"
}

func newTags() Tag {
	t := Tag{}
	t.tags = make([]string, 7)
	return t
}

func (t *Tag) append(k, v string) {
	t.tags = append(t.tags, k+":"+v)
}
