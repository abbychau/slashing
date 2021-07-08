package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/acme/autocert"
)

func main() {

	//Configurations
	configFileName := ""
	if len(os.Args) == 2 && fileExists(os.Args[1]) {
		configFileName = os.Args[1]
	} else {
		fmt.Println("Error: Config file does not exist. \nUsage: ./slashing config.txt")
		os.Exit(1)
	}
	file, err := os.Open(configFileName)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	domains := []string{}
	paths := map[string]string{}
	target := ""
	for scanner.Scan() {
		if target == "" {
			target = strings.Trim(scanner.Text(), " \t\r\n")
			continue
		}
		line := strings.Trim(scanner.Text(), " \t\r\n")
		if line != "" {
			parts := strings.Split(line, ":")
			domains = append(domains, parts[0])
			paths[parts[0]] = parts[1]
		}
	}

	//AutoCert Manager
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
		Cache:      autocert.DirCache(cacheDir()),
	}
	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate:           certManager.GetCertificate, //Cert generation
			PreferServerCipherSuites: true,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, // Required by Go (and HTTP/2 RFC), even if you only present ECDSA certs
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
			},
			//MinVersion:             tls.VersionTLS12,
			//CurvePreferences:       []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		},
	}

	// PROXY ROUTE
	origin, _ := url.Parse(target)
	director := func(req *http.Request) {
		// req.Header.Add("X-Forwarded-Host", req.Host)
		// req.Header.Add("X-Origin-Host", origin.Host)
		// req.Header.Add("X-Forwarded-For", req.Header.Get("X-Forwarded-For") ) // Forward Real IP?
		req.URL.Scheme = "http"
		req.URL.Host = origin.Host
	}
	proxy := &httputil.ReverseProxy{Director: director}

	// HTTP HANDLERS
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Incoming HTTP at ", r)
		possibleStaticFile := filepath.Join(paths[r.URL.Host], r.URL.Path)
		if fileExists(possibleStaticFile) {
			//File Exists
			http.ServeFile(w, r, possibleStaticFile)

		} else {
			//File not Exist
			//Do proxying
			// w.Header().Set("Strict-Transport-Security", "max-age=15768000 ; includeSubDomains")
			proxy.ServeHTTP(w, r)
		}

	})
	// http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("static"))))

	// SERVE, REDIRECT AUTO to HTTPS
	go func() {
		http.ListenAndServe(":http", certManager.HTTPHandler(nil))
	}()
	log.Fatal(server.ListenAndServeTLS("", "")) // SERVE HTTPS!

}

// cacheDir creates and returns a tempory cert directory under current path
func cacheDir() (dir string) {
	if u, _ := user.Current(); u != nil {
		fmt.Println(os.TempDir())
		//dir = filepath.Join(os.TempDir(), "cache-golang-autocert-"+u.Username)
		dir = filepath.Join(".", "cache-golang-autocert-"+u.Username)
		fmt.Println("Should be saving cache-go-lang-autocert-u.username to: ")
		fmt.Println(dir)
		if err := os.MkdirAll(dir, 0700); err == nil {
			return dir
		}
	}
	return ""
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
