package lib

import (
	"errors"
	"fmt"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"gopkg.in/mcuadros/go-syslog.v2/format"
	"strconv"
	"strings"
)

type (
	SyslogParser struct {
		_dirty   _logSlice
		logger   Logger
		config   ParserConfig
		template Template
	}

	Log struct {
		_time
		_request
		_response
		_upstream
		_http
		_connection
		bytesSent int
		_clientInfo
	}

	_time struct {
		timeLocal string
		msec      string
	}

	_request struct {
		remoteAddr     string
		serverProtocol string
		requestMethod  string
		host           string
		uri            string
		args           string
		requestTime    string

		_splitUri _splitUri
	}

	_response struct {
		status        int
		bodyBytesSent int
	}

	_upstream struct {
		upstreamResponseTime string
		upstreamAddr         string
		upstreamStatus       string
	}

	_http struct {
		httpReferer       string
		httpVia           string
		httpXForwardedFor string
		httpUserAgent     string
		sentHttpXProfile  string
	}

	_connection struct {
		connectionRequests int
		connection         string
	}

	_logSlice struct {
		client    string
		content   string
		tag       string
		facility  int
		hostname  string
		priority  int
		severity  int
		timestamp string
		tlsPeer   string
	}

	_splitUri struct {
		channel string
		quality string
		index   string
		prefix  string
	}

	_clientInfo struct {
		client   string
		tag      string
		hostname string
	}

	ParserConfig struct {
		PartsDelim  string
		StreamDelim string
		Template    Template
	}
)

func NewSyslogParser(logger Logger, config ParserConfig) SyslogParser {
	return SyslogParser{
		logger:   logger,
		config:   config,
		template: config.Template,
	}
}

func (s SyslogParser) Parse(parts format.LogParts) (Log, error) {
	s._dirty = s.toSlice(parts)

	if s._dirty.size() == 0 {
		return Log{}, errors.New(withMessage(constants.NOT_RECOGNIZE_LOGS, s._dirty.content))
	}

	nginxLogs := strings.Split(s._dirty.content, s.config.PartsDelim)

	if s.logger.IsDevelopment() {
		for k, v := range nginxLogs {
			s.logger.InfoLog(fmt.Sprintf("%d => %v", k, v))
		}
	}

	valueOf := s.template.makeClosure(nginxLogs)

	_req := _request{
		host:           valueOf("host"),
		remoteAddr:     valueOf("remote_addr"),
		uri:            valueOf("uri"),
		serverProtocol: "",
		requestMethod:  "",
		args:           "",
		requestTime:    "",
	}

	// @see readme
	streamUri := splitUri(valueOf("uri"), s.config.StreamDelim)
	_req._splitUri = s.streamParts(streamUri)

	return Log{
		_time: _time{
			timeLocal: "",
			msec:      "",
		},
		_request: _req,
		_response: _response{
			status:        0,
			bodyBytesSent: 0,
		},
		_upstream: _upstream{
			upstreamResponseTime: "",
			upstreamAddr:         "",
			upstreamStatus:       "",
		},
		_http: _http{
			httpReferer:       getIf(valueOf("http_referer")),
			httpVia:           getIf(valueOf("http_via")),
			httpXForwardedFor: getIf(valueOf("http_x_forwarded_for")),
			httpUserAgent:     getIf(valueOf("http_user_agent")),
			sentHttpXProfile:  getIf(valueOf("sent_http_x_profile")),
		},
		_connection: _connection{
			connectionRequests: strToInt(valueOf("connection_requests")),
			connection:         valueOf("connection"),
		},
		bytesSent: strToInt(valueOf("bytes_sent")),
		_clientInfo: _clientInfo{
			client:   s._dirty.client,
			tag:      s._dirty.tag,
			hostname: s._dirty.hostname,
		},
	}, nil
}

func (s SyslogParser) toSlice(parts format.LogParts) _logSlice {
	return _logSlice{
		client:   ifaceToStr(parts["client"]),
		content:  ifaceToStr(parts["content"]),
		tag:      ifaceToStr(parts["tag"]),
		hostname: ifaceToStr(parts["hostname"]),
	}
}

func (s SyslogParser) isStreamUri(splitUri []string) bool {
	if len(splitUri) == constants.LEN_STREAM_PARTS {
		return true
	}

	return false
}

func (s SyslogParser) defaultStreamUri() _splitUri {
	return _splitUri{
		channel: constants.UNKNOWN,
		quality: constants.UNKNOWN,
		index:   constants.UNKNOWN,
		prefix:  constants.UNKNOWN,
	}
}

func (s SyslogParser) streamParts(stream []string) _splitUri {
	// /streaming/muztv/324/vl2w/segment-1597220444-01972046.ts
	if isInetraTranscoder(len(stream)) {
		return _splitUri{
			prefix:  stream[1],
			channel: stream[2],
			quality: stream[4],
			index:   stream[5],
		}
	}

	// /streaming/karusel/324/variable.m3u8
	if isInetraMultibitrate(len(stream)) {
		return _splitUri{
			prefix:  constants.UNKNOWN,
			channel: stream[2],
			quality: constants.UNKNOWN,
			index:   stream[4],
		}
	}

	// /domashniy/tracks-v1a1/2020/08/13/11/38/56-06000.ts
	if isFlussonicTranscoder(len(stream)) {
		quality := s.splitQuality(stream[2])
		return _splitUri{
			prefix:  stream[1],
			channel: stream[1],
			quality: quality,
			index:   stream[8],
		}
	}

	// /karusel/tracks-v1a1/mono.m3u8
	if isFlussonicPlaylist(len(stream)) {
		quality := s.splitQuality(stream[2])
		return _splitUri{
			prefix:  constants.UNKNOWN,
			channel: stream[1],
			quality: quality,
			index:   stream[3],
		}
	}

	// /karusel/index.m3u8
	if isFlussonicMultibitrate(len(stream)) {
		return _splitUri{
			prefix:  constants.UNKNOWN,
			channel: stream[1],
			quality: constants.UNKNOWN,
			index:   stream[2],
		}
	}

	return s.defaultStreamUri()
}

// tracks-v1a1
func (s SyslogParser) splitQuality(quality string) string {
	q := strings.Split(quality, "-")
	if len(q) == 2 {
		return q[1]
	}
	return constants.UNKNOWN
}

// export getters

func (l Log) GetConnections() int {
	return l.connectionRequests
}

func (l Log) GetBytesSent() int {
	return l.bytesSent
}

func (l Log) GetStreamingServer() string {
	return l.host
}

func (l Log) GetQuality() string {
	return l._splitUri.Quality()
}

func (l Log) GetChannel() string {
	return l._splitUri.Channel()
}

func (l Log) GetRemoteAddr() string {
	return l.remoteAddr
}

func (l Log) GetUserAgent() string {
	return l.httpUserAgent
}

func (l Log) GetClientAddr() string {
	if len(l.client) == 0 {
		return constants.UNKNOWN
	}

	if strings.Contains(l.client, ":") {
		list := strings.Split(l.client, ":")

		if len(list) > 0 {
			return list[0]
		}
	}

	return l.client
}

// _logSlice methods

func (sl _logSlice) getContent() string {
	return sl.content
}

func (sl _logSlice) size() int {
	return len(sl.getContent())
}

// _splitUri methods

func (sp _splitUri) Channel() string {
	return sp.channel
}

func (sp _splitUri) Quality() string {
	return sp.quality
}

// safe methods

func ifaceToStr(value interface{}) string {
	if value == nil {
		return ""
	}

	return value.(string)
}

func strToInt(value string) int {
	converted, err := strconv.Atoi(value)

	if err != nil {
		return 0
	}

	return converted
}

func splitUri(uri string, delim string) []string {
	return strings.Split(uri, delim)
}

func getIf(value string) string {
	if len(value) == 0 || value == constants.EMPTY_VALUE {
		return constants.UNKNOWN
	}

	return value
}

func withMessage(message, body string) string {
	return fmt.Sprintf("%s\nСодержимое:\n%s", message, body)
}
