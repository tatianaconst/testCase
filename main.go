package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sevlyar/go-daemon"
)

const timeSample = "2006-01-02T15:04:05Z"

type record struct {
	reqTime time.Time
	address string
}

var db []record
var mu sync.Mutex

func handler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()

	db = append(db, record{time.Now(), r.Host})

	var str strings.Builder
	for k, v := range db {
		fmt.Fprintf(&str, "%d\t%s\t%s\n", k, v.reqTime.Format(timeSample), v.address)
	}
	_, err := w.Write([]byte(str.String()))
	if err != nil {
		log.Print("can't send response:", err)
	}

	mu.Unlock()

}

func main() {
//Make daemon
	cntxt := &daemon.Context{
        PidFileName: "sample.pid",
        PidFilePerm: 0644,
        LogFileName: "sample.log",
        LogFilePerm: 0640,
        WorkDir:     "./",
        Umask:       027,
        Args:        []string{"[go-daemon sample]"},
    }
	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("unable to run: ", err)
	}
	if d != nil {
		return
	}
	defer cntxt.Release()
    fmt.Print("- - - - - - - - - - - - - - -")
    fmt.Print("daemon started")

// Start ticker to clear db
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			mu.Lock()
			db = nil
			mu.Unlock()
		}
	}()
//Start listening
	http.HandleFunc("/", handler)
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatalf("can't start server: %s\n", err)
	}
}
