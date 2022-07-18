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

func init() {
	functions.HTTP("GoMercury", GoMercury)
}

func GoMercury(w http.ResponseWriter, r *http.Request) {
	var d MercuryRequest
	var m MercuryResponse

	m.Output = make(map[string]string)

	//Capture byte array input.
	var byteArray, err = io.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintln(w, html.EscapeString(err.Error()))
		return
	}

	//Convert byte array to string.
	var input = string(byteArray)

	switch {
	//If input is valid IP, use GeoIP.
	case net.ParseIP(input) != nil:
		d.IpAddress = input
		geoIp(d.IpAddress, &m)
		break
	//If input is valid Domain, use WhoIs.
	case validator.ValidateDomainByResolvingIt(input) == nil:
		d.Domain = input
		whoIs(d.Domain, &m)
		break
	//Return error to user.
	default:
		fmt.Fprintln(w, html.EscapeString("Not a valid IP Address or Domain Name."))
		return
	}

	//Output html of data inside MercuryResponse.
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
	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		m.Output["GeoIpErrorOnOpen"] = err.Error()
		return
	}
	defer db.Close()

	record, err := db.City(net.ParseIP(ipAddress))
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
