package function

import (
	"net/http"
	"net/url"
)

func checkRequest(r *http.Request, m *MercuryResponse, geoIPData []byte, w http.ResponseWriter) (u *url.URL) {
	//Check if we have sane GET request else bail
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
