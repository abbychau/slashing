package rdbms

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"slashing/utils"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// dbPath creates and returns a directory under current path
func dbPath() string {

	if u, _ := user.Current(); u != nil {
		path := filepath.Join(utils.CacheDir("cache-rdbms"), "rdbms-state.db")
		if !utils.FileExists(path) {
			if file, err := os.Create(path); err == nil {
				file.Close()
				return path
			}
		}

	}
	return ""
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func ListenAndServeHTTPServer(address string) *http.Server {
	db, err := sql.Open("sqlite3", dbPath())
	var mu sync.Mutex

	checkErr(err)

	router := http.NewServeMux()
	router.HandleFunc("/query", func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()
		timeStart := time.Now()
		rows, err := db.Query(req.PostFormValue("query"))

		timeElapsed := time.Since(timeStart)
		if err != nil {
			w.Write([]byte("Query Error"))
			return
		}

		b, err := json.Marshal(Response{Results: rows, Time: float64(timeElapsed / time.Millisecond)})
		if err != nil {
			w.Write([]byte("JSON serialization Error"))
			return
		}
		w.Write(b)
		rows.Close()
		mu.Unlock()
	})
	router.HandleFunc("/execute", func(res http.ResponseWriter, req *http.Request) {

	})
	s := &http.Server{
		Addr:           address,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return s
}

// Response represents a response from the HTTP service.
type Response struct {
	Results interface{} `json:"results,omitempty"`
	Time    float64     `json:"time,omitempty"` //in millisecond
}
