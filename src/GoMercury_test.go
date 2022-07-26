package function

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGoMercury(t *testing.T) {
	tests := []struct {
		body string
		want string
	}{
		{body: `81.2.69.134`, want: "GeoIPCity : Long Buckby"},
		{body: `81.2.69.1234`, want: "Not a valid IP Address or Domain Name.\n"},
		{body: `ebayzzzz.com`, want: "WhoisSuccessfulResponse : No match for &#34;EBAYZZZZ.COM&#34"},
	}
	for _, test := range tests {
		req := httptest.NewRequest("POST", "/", strings.NewReader(test.body))
		req.Header.Add("Content-Type", "text/plain")

		rr := httptest.NewRecorder()
		GoMercury(rr, req)

		if got := rr.Body.String(); !strings.Contains(got, test.want) {
			t.Errorf("GoMercury(%q) = %q, want %q", test.body, got, test.want)
		}
	}
}
