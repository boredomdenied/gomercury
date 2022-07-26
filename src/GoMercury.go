package function

import (
	"encoding/json"
	"net"
	"net/http"
	"sync"

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

func GoMercury(w http.ResponseWriter, r *http.Request) {
	//Create structs and initialize map
	var d MercuryRequest
	var m MercuryResponse

	m.Output = make(map[string]string)

	//Initialize Database
	geoIPData := initDatabase()

	//Sanity check the request
	u := checkRequest(r, &m, geoIPData, w)

	//Get Query Parameter
	var domain = u.Query().Get("domain")
	var IPaddress = u.Query().Get("ipaddress")

	switch {
	//If input is valid IP, use GeoIP
	case net.ParseIP(IPaddress) != nil:
		d.IpAddress = IPaddress
		geoIp(d.IpAddress, &m, geoIPData, w)
	//If input is valid Domain, use WhoIs
	case validator.ValidateDomainByResolvingIt(domain) == nil:
		d.Domain = domain
		whoIs(d.Domain, &m, w)
	default: //No valid IP or domain
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		m.Output["ErrorOnRequestParameter"] = "Not a Valid IP or Domain."
	}

	//Return Json Response
	json.NewEncoder(w).Encode(m.Output)
}
