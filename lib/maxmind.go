package lib

import (
	"fmt"
	"github.com/LimeHD/limehd-syslog-server/constants"
	"github.com/oschwald/geoip2-golang"
	"net"
)

type (
	GeoFinder struct {
		reader    *geoip2.Reader
		asnReader *geoip2.Reader
		_logger   Logger
	}

	GeoFinderConfig struct {
		MmdbPath    string
		AsnMmdbPath string
		Logger      Logger
	}

	_geoIdentity struct {
		isoName   string
		geoNameId uint
	}

	_city    _geoIdentity
	_country struct {
		_geoIdentity
		isoCode string
	}
	_asn struct {
		org    string
		number uint
	}

	GeoFinderResult struct {
		city    _city
		country _country
		asn     _asn
	}
)

func NewGeoFinder(config GeoFinderConfig) (GeoFinder, error) {
	var err error

	g := GeoFinder{}
	g.reader, err = g.openDatabase(config.MmdbPath)
	g._logger = config.Logger

	if err != nil {
		return GeoFinder{}, err
	}

	g.asnReader, err = g.openDatabase(config.AsnMmdbPath)

	if err != nil {
		return GeoFinder{}, err
	}

	if g._logger.IsDevelopment() {
		g._logger.InfoLog(fmt.Sprintf("Read MaxMind database from %s", config.MmdbPath))
	}

	return g, nil
}

func (g GeoFinder) Find(ip string) (*GeoFinderResult, error) {
	_ip := net.ParseIP(ip)
	record, err := g.reader.City(_ip)

	if err != nil {
		return nil, err
	}

	asnRecord, err := g.asnReader.ASN(_ip)

	if err != nil {
		return nil, err
	}

	return &GeoFinderResult{
		city: _city{
			isoName:   record.City.Names["en"],
			geoNameId: record.City.GeoNameID,
		},
		country: _country{
			_geoIdentity: _geoIdentity{
				isoName:   record.Country.Names["en"],
				geoNameId: record.Country.GeoNameID,
			},
			isoCode: record.Country.IsoCode,
		},
		asn: _asn{
			org:    asnRecord.AutonomousSystemOrganization,
			number: asnRecord.AutonomousSystemNumber,
		},
	}, nil
}

//

func (r GeoFinderResult) GetCountryGeoId() uint {
	return r.country.geoNameId
}

func (r GeoFinderResult) GetCountryName() string {
	if len(r.country.isoName) == 0 {
		return constants.UNKNOWN
	}

	return r.country.isoName
}

func (r GeoFinderResult) GetCountryIsoCode() string {
	if len(r.country.isoCode) == 0 {
		return constants.UNKNOWN
	}

	return r.country.isoCode
}

func (r GeoFinderResult) GetCityGeoId() uint {
	return r.city.geoNameId
}

func (r GeoFinderResult) GetCityName() string {
	return r.city.isoName
}

func (r GeoFinderResult) GetOrganization() string {
	if len(r.asn.org) == 0 {
		return constants.UNKNOWN
	}

	return r.asn.org
}

func (r GeoFinderResult) GetOrganizationNumber() uint {
	return r.asn.number
}

//

func (g GeoFinder) openDatabase(mmdbPath string) (*geoip2.Reader, error) {
	return geoip2.Open(mmdbPath)
}

func (g GeoFinder) Close() {
	_ = g.reader.Close()
	_ = g.asnReader.Close()
}

func (g GeoFinder) CloseMessage() string {
	return "Close MaxMind"
}
