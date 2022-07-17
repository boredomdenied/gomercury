package function

import (
	"fmt"
	"html"
	"io"
	"net"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/dchest/validator"
	"github.com/domainr/whois"
	"github.com/oschwald/geoip2-golang"
)

type MercuryRequest struct {
	Domain    string
	IpAddress string
}

type MercuryResponse struct {
	Output map[string]string
}

var input = ""

func init() {
	functions.HTTP("GoMercury", goMercury)
}

func goMercury(w http.ResponseWriter, r *http.Request) {
	var d MercuryRequest
	var m MercuryResponse
	m.Output = make(map[string]string)

	//Capture byte array input.
	var byteArray, err = io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	//Convert byte array to string.
	input = string(byteArray)

	//Check if input is an IP address.
	if net.ParseIP(input) != nil {
		d.IpAddress = input
		whoIs(d.IpAddress, &m)
	}

	//Check if input is a domain.
	if validator.IsValidDomain(input) {
		d.Domain = input
		geoIp(d.Domain, &m)
	}

	for k, v := range m.Output {
		fmt.Fprintln(w, html.EscapeString(k+" : "+v))
	}

}

func whoIs(domain string, m *MercuryResponse) {
	request, err := whois.NewRequest(domain)
	if err != nil {
		m.Output["WhoIsErrorOnRequest"] = err.Error()
		return
	}

	response, err := whois.DefaultClient.Fetch(request)
	if err != nil {
		m.Output["WhoIsErrorOnResponse"] = err.Error()
		return
	}

	m.Output["WhoisSuccessfulResponse"] = string(response.Body)
}

func geoIp(ipAddress string, m *MercuryResponse) {
	db, err := geoip2.Open("GeoIP2-City.mmdb")
	if err != nil {
		m.Output["GeoIpErrorOnOpen"] = err.Error()
		return
	}
	defer db.Close()

	record, err := db.City(net.IP(ipAddress))
	if err != nil {
		m.Output["GeoIpErrorForRecord"] = err.Error()
		return
	}
	m.Output["GeoIPCity"] = record.City.Names["en"]
	m.Output["GeoIPCountry"] = record.Country.Names["en"]
	m.Output["GeoIPContinent"] = record.Continent.Names["en"]
	m.Output["GeoIPLocationLatitude"] = fmt.Sprintf("%f", record.Location.Latitude)
	m.Output["GeoIPLocationLongitude"] = fmt.Sprintf("%f", record.Location.Longitude)

}
