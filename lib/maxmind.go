package lib

import (
	"fmt"
	"github.com/oschwald/geoip2-golang"
	"net"
)

type (
	GeoFinder struct {
		reader  *geoip2.Reader
		_logger Logger
	}

	GeoFinderConfig struct {
		MmdbPath string
		Logger   Logger
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

	GeoFinderResult struct {
		city    _city
		country _country
	}
)

func NewGeoFinder(config GeoFinderConfig) (GeoFinder, error) {
	var err error

	g := GeoFinder{}
	g.reader, err = g._openDabase(config.MmdbPath)
	g._logger = config.Logger

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
	}, nil
}

//

func (r GeoFinderResult) GetCountryGeoId() uint {
	return r.country.geoNameId
}

func (r GeoFinderResult) GetCountryName() string {
	return r.country.isoName
}

func (r GeoFinderResult) GetCountryIsoCode() string {
	return r.country.isoCode
}

func (r GeoFinderResult) GetCityGeoId() uint {
	return r.city.geoNameId
}

func (r GeoFinderResult) GetCityName() string {
	return r.city.isoName
}

//

func (g GeoFinder) _openDabase(mmdbPath string) (*geoip2.Reader, error) {
	return geoip2.Open(mmdbPath)
}

func (g GeoFinder) Close() {
	_ = g.reader.Close()
}
