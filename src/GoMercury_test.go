package function

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGoMercury(t *testing.T) {
	tests := []struct {
		domain    string
		ipaddress string
		want      string
		method    string
	}{
		{method: `GET`, ipaddress: `81.2.69.134`, want: `{"GeoIPCity":"Long Buckby","GeoIPContinent":"Europe","GeoIPCountry"`},
		{method: `GET`, ipaddress: `81.2.69.1234`, want: `{"ErrorOnRequestParameter":"Not a Valid IP or Domain."}`},
		{method: `GET`, domain: `ebayzzzz.com`, want: `{"WhoIsErrorOnResolution":"Domain resolution unsuccessful."}`},
		{method: `POST`, domain: `ebay.com`, want: `{"ErrorOnMethodType":"Not a Valid GET request."`},
		{method: `GET`, domain: `ebay.com`, want: `{"WhoisSuccessfulResponse":"   Domain Name: EBAY.COM"}`},
	}

	for _, test := range tests {
		var queryparam, testLocation = "", ""

		if test.domain != "" {
			queryparam = "/?domain=" + test.domain
			testLocation = test.domain
		}

		if test.ipaddress != "" {
			queryparam = "/?ipaddress=" + test.ipaddress
			testLocation = test.ipaddress
		}

		req := httptest.NewRequest(test.method, queryparam, nil)
		req.Header.Add("Content-Type", "text/plain")

		rr := httptest.NewRecorder()
		GoMercury(rr, req)

		if got := rr.Body.String(); !strings.Contains(got, test.want) {
			t.Errorf("GoMercury(%q) = %q, want %q", testLocation, got, test.want)
		}
	}
}
