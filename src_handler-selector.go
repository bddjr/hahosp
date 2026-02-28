package hahosp

import (
	"net/http"

	hahosp_utils "github.com/bddjr/hahosp/utils"
)

type HandlerSelector struct {
	// Handles on HTTPS.
	HTTPS http.Handler

	// Handles on HTTP.
	// If nil, redirect to HTTPS.
	HTTP http.Handler
}

func (hs *HandlerSelector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.TLS != nil {
		// HTTPS
		if h := hs.HTTPS; h != nil {
			h.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	} else {
		// HTTP
		if h := hs.HTTP; h != nil {
			h.ServeHTTP(w, r)
		} else {
			hahosp_utils.RedirectToHttps_ForceSamePort(w, r, 307)
		}
	}
}
