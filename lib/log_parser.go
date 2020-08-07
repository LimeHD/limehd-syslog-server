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
		_dirty _logSlice
		logger Logger
		config ParserConfig
	}

	Log struct {
		_time
		_request
		_response
		_upstream
		_http
		_connection
		bytesSent int
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

	ParserConfig struct {
		PartsDelim  string
		StreamDelim string
	}
)

func NewSyslogParser(logger Logger, config ParserConfig) SyslogParser {
	return SyslogParser{
		logger: logger,
		config: config,
	}
}

func (s SyslogParser) Parse(parts format.LogParts) (Log, error) {
	s._dirty = s.toSlice(parts)

	if s._dirty.size() == 0 {
		return Log{}, errors.New("It seems to be missing any logs")
	}

	_logFormatParts := strings.Split(s._dirty.content, s.config.PartsDelim)

	if s.logger.IsDevelopment() {
		for k, v := range _logFormatParts {
			s.logger.InfoLog(fmt.Sprintf("%d => %v", k, v))
		}
	}

	_req := _request{
		host:           _logFormatParts[constants.POS_HOST],
		remoteAddr:     _logFormatParts[constants.POS_REMOTE_ADDR],
		uri:            _logFormatParts[constants.POS_URI],
		serverProtocol: "",
		requestMethod:  "",
		args:           "",
		requestTime:    "",
		_splitUri:      s.defaultStreamUri(),
	}

	// example: /streaming/domashniy/324/vh1w/playlist.m3u8
	__splitUri := _safeSplitUri(_logFormatParts[constants.POS_URI], s.config.StreamDelim)

	if s.isStreamUri(__splitUri) {
		_req._splitUri = _splitUri{
			prefix:  __splitUri[1],
			channel: __splitUri[2],
			quality: __splitUri[4],
			index:   __splitUri[5],
		}
	}

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
			httpReferer:       _getOrUnknown(_logFormatParts[constants.POS_HTTP_REFERER]),
			httpVia:           _getOrUnknown(_logFormatParts[constants.POS_HTTP_VIA]),
			httpXForwardedFor: _getOrUnknown(_logFormatParts[constants.POS_HTTP_XFORWARD]),
			httpUserAgent:     _getOrUnknown(_logFormatParts[constants.POS_HTTP_USER_AGENT]),
			sentHttpXProfile:  _getOrUnknown(_logFormatParts[constants.POST_HTTP_SENT_X]),
		},
		_connection: _connection{
			connectionRequests: _safeStringToInt(_logFormatParts[constants.POS_CONNECTION_REQUESTS]),
			connection:         _logFormatParts[constants.POS_CONNECTIONS],
		},
		bytesSent: _safeStringToInt(_logFormatParts[constants.POS_BYTES_SENT]),
	}, nil
}

func (s SyslogParser) toSlice(parts format.LogParts) _logSlice {
	return _logSlice{
		client:   _safeInterfaceToString(parts["client"]),
		content:  _safeInterfaceToString(parts["content"]),
		tag:      _safeInterfaceToString(parts["tag"]),
		hostname: _safeInterfaceToString(parts["hostname"]),
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

func _safeInterfaceToString(value interface{}) string {
	if value == nil {
		return ""
	}

	return value.(string)
}

func _safeStringToInt(value string) int {
	converted, err := strconv.Atoi(value)

	if err != nil {
		return 0
	}

	return converted
}

func _safeSplitUri(uri string, delim string) []string {
	return strings.Split(uri, delim)
}

func _getOrUnknown(value string) string {
	if len(value) == 0 || value == constants.EMPTY_VALUE {
		return constants.UNKNOWN
	}

	return value
}