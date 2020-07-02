package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// This is in our opinion the best config for a server to be compatible, modern and secure at the same time
// scoring A+ in SSL Labs, IPv6 and HTTP2 with very quick response (have a nice time !!!)

const (
	globalroot   = "/var/www/html"
	firstpage    = "index"
	notfoundfile = "404.html" // inside its own domain's folder
)

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

	go log.Fatal(http.ListenAndServe(":80", certManager.HTTPHandler(nil)))
	log.Fatal(server.ListenAndServeTLS("", ""))
}

// web server for root
func root(w http.ResponseWriter, req *http.Request) {
	var out string
	var notfound bool

	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	out = fmt.Sprintf("Host: %s\n", req.Host)
	for k, v := range req.Header {
		out += fmt.Sprintf("%s : %s\n", k, v[0])
	}
	fmt.Printf(out)

	rootdir := fmt.Sprintf("%s/%s/", globalroot, req.Host) // "/var/www/html/domain.com/"
	namefile := strings.TrimRight(rootdir+req.URL.Path[1:], "/")
	fileinfo, err := os.Stat(namefile)
	if err != nil {
		// file does not exist
		namefile = fmt.Sprintf("%s%s", rootdir, notfoundfile)
		notfound = true
	} else if fileinfo.IsDir() {
		// it is a folder, get in and access the index page
		namefile = namefile + "/" + firstpage + ".html"
		_, err := os.Stat(namefile)
		if err != nil {
			// no index file inside this folder
			namefile = fmt.Sprintf("%s%s", rootdir, notfoundfile)
			notfound = true
		}
	}
	fmt.Println("File:", namefile)
	if notfound {
		fileinfo, err = os.Stat(namefile)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			return
		}
	}
	fr, err := os.Open(namefile)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return
	}
	defer fr.Close()

	http.ServeContent(w, req, namefile, fileinfo.ModTime(), fr)
}
