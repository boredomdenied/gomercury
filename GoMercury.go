package function

import (
	"context"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"

	"cloud.google.com/go/storage"
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

var (
	geoIPData     []byte
	loadGeoIpOnce sync.Once
)

func init() {
	functions.HTTP("GoMercury", GoMercury)
}

func onceBody() []byte {
	loadGeoIpOnce.Do(func() {
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Fatalf("Error opening storage.NewClient: %s.", err)
			return
		}
		rc, err := client.Bucket("gomercury-bucket356415").Object("GeoLite2-City.mmdb").NewReader(ctx)
		if err != nil {
			log.Fatalf("Error opening storage bucket: %s.", err)
			return
		}
		defer rc.Close()
		geoIPData, err = io.ReadAll(rc)
		if err != nil {
			log.Fatalf("Error with reading geoIPData: %s.", err)
			return
		}
	})
	return geoIPData
}

func GoMercury(w http.ResponseWriter, r *http.Request) {
	var d MercuryRequest
	var m MercuryResponse

	m.Output = make(map[string]string)

	geoIPData := onceBody()

	//Check if we have sane GET request else bail.
	u, err := url.Parse(r.RequestURI)
	if err != nil {
		fmt.Fprintln(w, html.EscapeString(err.Error()))
		return
	}
	if r.Method != "GET" {
		fmt.Fprintln(w, html.EscapeString("Not a GET request"))
		return
	}

	// Input is the first query
	var input = u.RawQuery

	switch {
	//If input is valid IP, use GeoIP.
	case net.ParseIP(input) != nil:
		d.IpAddress = input
		geoIp(d.IpAddress, &m, geoIPData)
	//If input is valid Domain, use WhoIs.
	case validator.ValidateDomainByResolvingIt(input) == nil:
		d.Domain = input
		whoIs(d.Domain, &m)
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

func geoIp(ipAddress string, m *MercuryResponse, geoIPData []byte) {
	db, err := geoip2.FromBytes(geoIPData)
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
