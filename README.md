# HTTPS And HTTP On Same Port

Listen HTTPS and HTTP on same port.

> If you only need redirect to HTTPS, recommended use [hlfhr](https://github.com/bddjr/hlfhr).

---

## Setup

```
go get github.com/bddjr/hahosp
```

```go
// test
srv := &http.Server{
    Addr:    ":5688"
    Handler: http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        if r.TLS != nil {
            io.WriteString(w, "You'r using HTTPS\n")
        } else {
            io.WriteString(w, "You'r using HTTP\n")
        }
    }),
}

// Use hahosp.ListenAndServe
err := hahosp.ListenAndServe(srv, "localhost.crt", "localhost.key")
```

```go
// redirect
srv := &http.Server{
    Addr:    ":5688"
    Handler: http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
        if r.TLS == nil {
            // HTTP
            url := "https://" + r.Host + r.URL.Path
            if r.URL.ForceQuery || r.URL.RawQuery != "" {
                url += "?" + r.URL.RawQuery
            }
            w.Header().Set("Location", url)
            w.WriteHeader(302)
            return
        }
        // HTTPS
        io.WriteString(w, "ok\n")
    }),
}

// Use hahosp.ListenAndServe
err := hahosp.ListenAndServe(srv, "localhost.crt", "localhost.key")
```


---

## Logic

```mermaid
flowchart TD
	VirtualListener("Hijacking net.Listener")
    VirtualAccept("🔄📥Virtual accept")
    Accept("🔄 Accept")
    Serve(["✅ serve..."])
	LooksLike{{"Read first byte looks like ?"}}
	Close(["❌ Close."])
    HijackingNetConn("Hijacking net.Conn")
    SentToVA(["📤Send to Virtual accept"])
    NewTLS("New tls.Conn")

    VirtualListener -- "http.Server Serve" --> VirtualAccept -- "async" --> Serve
    VirtualListener -- "async hahosp Serve" --> Accept -- "async"  --> HijackingNetConn --> LooksLike
    LooksLike -- "❓Unknown" --> Close
    LooksLike -- "📄HTTP" --> SentToVA
    LooksLike -- "🔐TLS" --> NewTLS --> SentToVA
```

---

## Test

```
git clone https://github.com/bddjr/hahosp
cd hahosp
chmod +x run.sh
./run.sh
```

---

## Reference

https://github.com/bddjr/hlfhr

---

## License

[BSD-3-clause license](LICENSE.txt)
