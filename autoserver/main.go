package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

// This is in our opinion the best config for a server to be compatible, modern and secure at the same time
// scoring A+ in SSL Labs, IPv6 and HTTP2 with very quick response (have a nice time !!!)

const (
	globalroot   = "/var/www/html"
	firstpage    = "index"
	notfoundfile = "404.html" // inside its own domain's folder
	logFile      = "/var/log/perfect/access.log"
)

// Logger : defines all the loggin system
type Logger struct {
	LogFile string      // log file to write
	InfoLog *log.Logger // info msg logger
}

var (
	logger *Logger
)

func init() {
	// starting the global logger system first of all
	logger = &Logger{}
	logger.LogFile = logFile
	file, err := os.OpenFile(logger.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Unable to open the error logging file:", err)
	}
	logger.InfoLog = log.New(io.MultiWriter(file, os.Stderr), "INFO:", log.Ldate|log.Ltime)
}

func main() {
	// setup a simple handler which sends a HTHS header for six months (!)
	http.HandleFunc("/", root)

	// look for the domains to be served from command line args
	flag.Parse()
	domains := flag.Args()
	if len(domains) == 0 {
		log.Fatalf("fatal; specify domains as arguments")
	}

	fmt.Printf("Serving HTTPS on domains:\n")
	for k, domain := range domains {
		fmt.Printf("[%d]\t%s\n", k, domain)
	}

	// create the autocert.Manager with domains and path to the cache
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
	}

	// optionally use a cache dir
	dir := cacheDir()
	if dir != "" {
		certManager.Cache = autocert.DirCache(dir)
	}

	// create the server itself
	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
			MinVersion:     tls.VersionTLS12,
		},
	}

	// log.Printf("Serving http/https for domains: %+v", domains)
	go func() {
		// serve HTTP, which will redirect automatically to HTTPS
		h := certManager.HTTPHandler(nil)
		log.Fatal(http.ListenAndServe(":http", h))
	}()

	// serve HTTPS!
	log.Fatal(server.ListenAndServeTLS("", ""))
}

// cacheDir makes a consistent cache directory inside /tmp. Returns "" on error.
func cacheDir() (dir string) {
	if u, _ := user.Current(); u != nil {
		dir = filepath.Join(os.TempDir(), "cache-golang-autocert-"+u.Username)
		if err := os.MkdirAll(dir, 0700); err == nil {
			return dir
		}
	}
	return ""
}

// web server for root
func root(w http.ResponseWriter, req *http.Request) {
	var notfound bool

	w.Header().Add("Strict-Transport-Security", "max-age=63072000; includeSubDomains")

	// out := fmt.Sprintf("Host: %s\n", req.Host)
	// for k, v := range req.Header {
	// 	out += fmt.Sprintf("%s : %s\n", k, v[0])
	// }
	// fmt.Printf(out)

	userAgent := strings.ToLower(req.Header.Get("User-Agent"))

	if strings.HasSuffix(req.URL.Path, ".zip") { // log the download
		logger.InfoLog.Printf("IP: [%s] File: [%s]\n", req.RemoteAddr, req.URL.Path)
	} else if strings.Contains(userAgent, "google") { // log the bot
		logger.InfoLog.Printf("IP: [%s] Host: [%s] Bot: [%s]\n", req.RemoteAddr, req.Host, userAgent)
	}

	// this part is only to serve Bots
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
	// fmt.Println("File:", namefile)
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
