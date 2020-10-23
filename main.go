package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var clientA *http.Client
var clientB *http.Client

func httpGet(client *http.Client, requestString string) {
	request, err := http.NewRequest("GET", requestString, nil)
	if err != nil {
		log.Printf("%v", err)
		return
	}

	_, err = client.Do(request)
	if err != nil {
		log.Printf("[%v]", err)
	}
}

func httpGetA() {
	httpGet(clientA, "http://0.0.0.0:5000")
}

func httpGetB() {
	httpGet(clientB, "http://0.0.0.0:6000")
}

func deafaultHandler(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte(``))
	if err != nil {
		log.Printf("%v", err)
	}
}

func withoutGoroutine(w http.ResponseWriter, r *http.Request) {

	httpGetA()

	_, err := w.Write([]byte(``))
	if err != nil {
		log.Printf("%v", err)
	}
}

func withGoroutine(w http.ResponseWriter, r *http.Request) {

	httpGetA()

	_, err := w.Write([]byte(``))
	if err != nil {
		log.Printf("%v", err)
	}

	go httpGetB()
}

func withSleepyGoroutine(w http.ResponseWriter, r *http.Request) {

	httpGetA()

	_, err := w.Write([]byte(``))
	if err != nil {
		log.Printf("%v", err)
	}

	go func() {
		time.Sleep(1 * time.Millisecond)
	}()
}

func newHTTPClient() *http.Client {
	idleTimeSeconds := 30
	timeoutSeconds := 2
	idlePerHost := 20
	maxConnsPerHost := 20

	tr := &http.Transport{
		MaxIdleConnsPerHost: maxConnsPerHost,
		MaxConnsPerHost:     idlePerHost,
		IdleConnTimeout:     time.Duration(idleTimeSeconds) * time.Second,
		DisableKeepAlives:   false,
	}

	return &http.Client{
		Transport: tr,
		Timeout:   time.Second * time.Duration(timeoutSeconds),
	}

}

func main() {

	clientA = newHTTPClient()
	clientB = newHTTPClient()

	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	router := mux.NewRouter()
	router.HandleFunc("/", deafaultHandler)
	router.HandleFunc("/withoutgoroutine", withoutGoroutine)
	router.HandleFunc("/withgoroutine", withGoroutine)
	router.HandleFunc("/withsleepygoroutine", withSleepyGoroutine)

	listenAddress := fmt.Sprintf(":%s", port)
	log.Printf("Listening on %s", listenAddress)

	log.Fatal(http.ListenAndServe(listenAddress, router))
}
