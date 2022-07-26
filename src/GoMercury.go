package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
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
	Output map[string]string `json:"output"`
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
	// geoIPData, err := os.ReadFile("./GeoLite2-City.mmdb")

	//Check if we have sane GET request else bail.
	u, err := url.Parse(r.RequestURI)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		m.Output["ErrorOnRequestParameter"] = err.Error()
	}
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("Content-Type", "application/json")
		m.Output["ErrorOnMethodType"] = "Not a Valid GET request."
	}

	var domain = u.Query().Get("domain")
	var IPaddress = u.Query().Get("ipaddress")

	switch {
	//If input is valid IP, use GeoIP.
	case net.ParseIP(IPaddress) != nil:
		d.IpAddress = IPaddress
		geoIp(d.IpAddress, &m, geoIPData, w)
	//If input is valid Domain, use WhoIs.
	case validator.ValidateDomainByResolvingIt(domain) == nil:
		d.Domain = domain
		whoIs(d.Domain, &m, w)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		m.Output["ErrorOnRequestParameter"] = "Not a Valid IP or Domain."
	}

	//Output html of data inside MercuryResponse.
	json.NewEncoder(w).Encode(m.Output)
}

func whoIs(domain string, m *MercuryResponse, w http.ResponseWriter) {

	//Request failure.
	request, err := whois.NewRequest(domain)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		m.Output["WhoIsErrorOnRequest"] = err.Error()
		return
	}

	//Fetch failure.
	response, err := whois.DefaultClient.Fetch(request)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		m.Output["WhoIsErrorOnFetch"] = err.Error()
		return
	}
	var noResolution = strings.Contains(string(response.Body), "No match for")

	//Domain resolution failure.
	if noResolution {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		m.Output["WhoIsErrorOnResolution"] = "Domain resolution unsuccessful."
		return
	}

	//Extract subset of successful response.
	var successResponse = strings.Split(string(response.Body), "\r")
	m.Output["WhoisSuccessfulResponse"] = successResponse[0]
}

func geoIp(ipAddress string, m *MercuryResponse, geoIPData []byte, w http.ResponseWriter) {

	//Database access failure.
	db, err := geoip2.FromBytes(geoIPData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		m.Output["GeoIpErrorOnOpen"] = err.Error()
		return
	}
	defer db.Close()

	//Ip address failure.
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
