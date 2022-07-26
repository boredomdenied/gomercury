package function

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/dchest/validator"
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

	//Initialize function database
	geoIPData := onceBody()

	//Sanity check on request
	u := CheckRequest(r, &m, geoIPData, w)

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

func CheckRequest(r *http.Request, m *MercuryResponse, geoIPData []byte, w http.ResponseWriter) (u *url.URL) {
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
	return
}
