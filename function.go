package function

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/domainr/whois"
)

func init() {
	functions.HTTP("GoMercury", goMercury)
}

func goMercury(w http.ResponseWriter, r *http.Request) {
	var d struct {
		Domain string `json:"domain"`
	}

	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		switch err {
		case io.EOF:
			fmt.Fprint(w, "First case!")
			return
		default:
			log.Printf("json.NewDecoder: %v", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
	}

	if d.Domain == "" {
		fmt.Fprint(w, "Second case!")
		return
	}

	request, err := whois.NewRequest(d.Domain)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	response, err := whois.DefaultClient.Fetch(request)
	if err != nil {
		fmt.Fprintln(w, err)
	}

	fmt.Fprintln(w, html.EscapeString(string(response.Body)))

}
