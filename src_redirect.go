package hahosp

import "net/http"

// Redirect without HTTP body.
func Redirect(w http.ResponseWriter, code int, url string) {
	w.Header()["Location"] = []string{url}
	w.WriteHeader(code)
}

// Redirect without HTTP body.
func RedirectToHttps(w http.ResponseWriter, r *http.Request, code int) {
	url := "https://" + r.Host + r.URL.Path
	if r.URL.ForceQuery || r.URL.RawQuery != "" {
		url += "?" + r.URL.RawQuery
	}
	Redirect(w, code, url)
}
