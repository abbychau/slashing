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
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func main() {
	log.Println("Start slashing...")
	targets, domains, paths := loadConfigurations()

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

	log.Println("Starting HTTP->HTTPS redirector and HTTPS server...")
	go func() {
		log.Fatal(http.ListenAndServe(":http", certManager.HTTPHandler(nil)))
	}()
	go func() {
		log.Fatal(server.ListenAndServeTLS("", ""))
	}()
	gracefulBlocker(server)
}

// cacheDir creates and returns a tempory cert directory under current path
func cacheDir() (dir string) {
	if u, _ := user.Current(); u != nil {
		//dir = filepath.Join(os.TempDir(), "cache-golang-autocert-"+u.Username)
		dir = filepath.Join(".", "cache-golang-autocert-"+u.Username)
		log.Printf("Certificate cache directory is : %v \n", dir)
		if err := os.MkdirAll(dir, 0700); err == nil {
			return dir
		}
	}
	return ""
}

func fileExists(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return false
	}
	// IsDir is short for fileInfo.Mode().IsDir()
	if fileInfo.IsDir() {
		// file is a directory
		return false
	}
	return true
}

func gracefulBlocker(srv *http.Server) {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown With Error: ", err)
	}

	log.Println("Server exiting")
}
func loadConfigurations() ([]string, []string, map[string]string) {
	//Configurations
	configFileName := ""
	if len(os.Args) == 2 && fileExists(os.Args[1]) {
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
	for scanner.Scan() {
		if len(targets) == 0 {
			targetsLine := strings.Trim(scanner.Text(), " \t\r\n")
			targets = strings.Split(targetsLine, ",")
			continue
		}
		line := strings.Trim(scanner.Text(), " \t\r\n")
		if line != "" {
			parts := strings.Split(line, ":")
			domains = append(domains, parts[0])
			paths[parts[0]] = parts[1]
		}
	}
	return targets, domains, paths
}

// func startWorkers()
//     cmd := exec.Command( "shell.sh" )
//     err := cmd.Start()
//     if err != nil {
//         return err
//     }
//     pid := cmd.Process.Pid
//     // use goroutine waiting, manage process
//     // this is important, otherwise the process becomes in S mode
//     go func() {
//         err = cmd.Wait()
//         fmt.Printf("Command finished with error: %v", err)
//     }()
//     return nil
// }
