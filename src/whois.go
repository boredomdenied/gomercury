package function

import (
	"net/http"
	"strings"

	"github.com/domainr/whois"
)

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
