package hahosp

import "net/http"

type HandlerSelector struct {
	// Handles on HTTPS.
	HTTPS http.Handler

	// Handles on HTTP.
	// If nil, redirect to HTTPS.
	HTTP http.Handler
}

func (this *HandlerSelector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.TLS != nil {
		// HTTPS
		if this.HTTPS != nil {
			this.HTTPS.ServeHTTP(w, r)
		} else {
			http.NotFound(w, r)
		}
	} else {
		// HTTP
		if this.HTTP != nil {
			this.HTTP.ServeHTTP(w, r)
		} else {
			RedirectToHttps(w, r, 302)
		}
	}
}
