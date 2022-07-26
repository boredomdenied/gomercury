package function

import (
	"fmt"
	"net"
	"net/http"

	"github.com/oschwald/geoip2-golang"
)

func geoIp(ipAddress string, m *MercuryResponse, geoIPData []byte, w http.ResponseWriter) {

	//Database access failure
	db, err := geoip2.FromBytes(geoIPData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		m.Output["GeoIpErrorOnOpen"] = err.Error()
		return
	}
	defer db.Close()

	//Ip address failure
	record, err := db.City(net.ParseIP(ipAddress))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		m.Output["GeoIpErrorForRecord"] = err.Error()
		return
	}

	m.Output["GeoIPCity"] = record.City.Names["en"]
	m.Output["GeoIPCountry"] = record.Country.Names["en"]
	m.Output["GeoIPContinent"] = record.Continent.Names["en"]
	m.Output["GeoIPLocationLatitude"] = fmt.Sprintf("%f", record.Location.Latitude)
	m.Output["GeoIPLocationLongitude"] = fmt.Sprintf("%f", record.Location.Longitude)
}
