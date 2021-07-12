package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"path/filepath"
	"slashing/rdbms"
	"slashing/redis"
	"slashing/utils"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func main() {
	log.Println("Start slashing...")

	targets, domains, paths, redisAddr, rdbmsAddr := loadConfigurations()
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domains...),
		Cache:      autocert.DirCache(utils.CacheDir("cache-autocert")),
	}
	httpServers := []*http.Server{}
	go func() {
		log.Println("Starting Redis server...")
		log.Fatal(redis.ListenAndServeRedisServer(redisAddr)) //Graceful shutdown not provided
	}()
	go func() {
		log.Println("Starting SQL HTTP server...")
		SQLHTTPServer := rdbms.ListenAndServeHTTPServer(rdbmsAddr)
		httpServers = append(httpServers, SQLHTTPServer)
		log.Fatal(SQLHTTPServer.ListenAndServe())
	}()
	if len(domains) > 0 {
		go func() {
			log.Println("Starting HTTP->HTTPS redirector and HTTPS server...")
			log.Fatal(http.ListenAndServe(":http", certManager.HTTPHandler(nil))) //Non-Graceful shutdown is not harmful
		}()
		TLSServer := getTLSServer(targets, domains, paths, &certManager)
		go func() {
			log.Println("Starting HTTPS server...")
			log.Fatal(TLSServer.ListenAndServeTLS("", ""))
			httpServers = append(httpServers, TLSServer)
		}()
	}
	gracefulBlocker(httpServers)
}

func gracefulBlocker(servers []*http.Server) {
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) //10 seconds should be enough, or else timeout
	defer cancel()
	for _, srv := range servers {
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown With Error: ", err)
		}
	}

	log.Println("Server exiting")
}

//loadConfigurations() returns targets, domains, paths, redis-port, rdbms-port
func loadConfigurations() ([]string, []string, map[string]string, string, string) {
	//Configurations
	configFileName := ""
	if len(os.Args) == 2 && utils.FileExists(os.Args[1]) {
		configFileName = os.Args[1]
	} else {
		log.Println("Error: Config file does not exist. \nUsage: ./slashing config.txt")
		os.Exit(1)
	}
	file, err := os.Open(configFileName)
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(file)
	domains := []string{}
	paths := map[string]string{}
	targets := []string{}
	var redis string
	var rdbms string
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " \t\r\n")
		if line == "" || string(line[0]) == "#" {
			continue
		}
		parts := strings.Split(line, "=")
		switch parts[0] {
		case "backend":
			targets = append(targets, parts[1])
		case "redis":
			redis = parts[1]
		case "rdbms":
			rdbms = parts[1]
		case "domain":
			valueParts := strings.Split(parts[1], ":")
			domains = append(domains, valueParts[0])
			paths[valueParts[0]] = valueParts[1]
		}

	}
	return targets, domains, paths, redis, rdbms
}

func getTLSServer(targets []string, domains []string, paths map[string]string, certManager *autocert.Manager) *http.Server {

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
	targetID := 0

	director := func(req *http.Request) {
		// req.Header.Add("X-Forwarded-Host", req.Host)
		// req.Header.Add("X-Origin-Host", origin.Host)
		// req.Header.Add("X-Forwarded-For", req.Header.Get("X-Forwarded-For") ) // Forward Real IP?
		if targetID == len(targets) {
			targetID = 0
		}
		req.URL.Scheme = "http"
		req.URL.Host = targets[targetID]
		targetID++
	}
	proxy := &httputil.ReverseProxy{Director: director}

	// HTTP HANDLERS
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Incoming HTTP:", r.Host, r.URL.Path)

		possibleStaticFile := filepath.Join(paths[r.Host], r.URL.Path)
		log.Println(possibleStaticFile)
		if utils.FileExists(possibleStaticFile) {
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

	return server
}
