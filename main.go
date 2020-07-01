package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/acme/autocert"
)

// This is in our opinion the best config for a server to be compatible, modern and secure at the same time
// scoring A+ in SSL Labs, IPv6 and HTTP2 with very quick response (have a nice time !!!)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", root)

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache("/etc/certs"),
	}

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	go http.ListenAndServe(":80", certManager.HTTPHandler(nil))
	log.Fatal(server.ListenAndServeTLS("", ""))
}

// web server for root
func root(w http.ResponseWriter, req *http.Request) {
	var out string

	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	out = fmt.Sprintf("Host: %s\n", req.Host)
	for k, v := range req.Header {
		out += fmt.Sprintf("%s : %s\n", k, v[0])
	}

	w.Write([]byte(out))
}
