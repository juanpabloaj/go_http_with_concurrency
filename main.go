package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/gops/agent"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

var jsonContent = ``

//var jsonLil = `{"key1":"val", "key2":{"deep":"blue"}}`
var jsonLil = `{"key1":"val"}`

var clientA *http.Client
var clientB *http.Client
var mqttClient mqtt.Client
var udpClient net.Conn
var metricChan chan []byte

func httpGet(ctx context.Context, client *http.Client, requestString string) {
	request, err := http.NewRequestWithContext(ctx, "GET", requestString, nil)
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
	ctx := context.Background()
	httpGet(ctx, clientA, "http://0.0.0.0:5000")
}

func httpGetB(ctx context.Context) {
	httpGet(ctx, clientB, "http://0.0.0.0:6000")
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

	go httpGetB(r.Context())
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

func withTelegrafGoroutine(w http.ResponseWriter, r *http.Request) {

	httpGetA()

	_, err := w.Write([]byte(``))
	if err != nil {
		log.Printf("%v", err)
	}

	go func() {
		udpClient.Write([]byte("m1,tag1=tag_value value=1"))
	}()

	go func() {
		udpClient.Write([]byte("m1,tag1=tag_value value=1"))
	}()
	//go func() {
	//	fmt.Fprintf(udpClient, "m1,tag1=tag_value value=2")
	//}()
}

func withTelegrafToChannel(w http.ResponseWriter, r *http.Request) {

	httpGetA()

	_, err := w.Write([]byte(``))
	if err != nil {
		log.Printf("%v", err)
	}

	metricChan <- []byte("m1,tag1=tag_value value=1")
}

func withMultiTelegrafToChannel(w http.ResponseWriter, r *http.Request) {

	httpGetA()

	_, err := w.Write([]byte(``))
	if err != nil {
		log.Printf("%v", err)
	}

	metricsNumber := 1
	if val := mux.Vars(r)["metrics_number"]; val != "" {
		n, _ := strconv.Atoi(val)
		if n > 0 {
			metricsNumber = n
		}
	}

	for i := 0; i < metricsNumber; i++ {
		metricChan <- []byte(jsonLil)
		time.Sleep(200 * time.Microsecond)
	}

}

func withMqttChannel(w http.ResponseWriter, r *http.Request) {

	httpGetA()

	_, err := w.Write([]byte(``))
	if err != nil {
		log.Printf("%v", err)
	}

	metricsNumber := 1
	if val := mux.Vars(r)["metrics_number"]; val != "" {
		n, _ := strconv.Atoi(val)
		if n > 0 {
			metricsNumber = n
		}
	}

	for i := 0; i < metricsNumber; i++ {
		metricChan <- []byte(jsonLil)
		time.Sleep(200 * time.Microsecond)
	}

}

func withMultiTelegrafJSON(w http.ResponseWriter, r *http.Request) {

	httpGetA()

	_, err := w.Write([]byte(``))
	if err != nil {
		log.Printf("%v", err)
	}
	metricsNumber := 1
	if val := os.Getenv("METRICS_NUMBER"); val != "" {
		n, _ := strconv.Atoi(val)
		if n > 0 {
			metricsNumber = n
		}
	}

	for i := 0; i < metricsNumber; i++ {
		metricChan <- []byte(jsonContent)
		time.Sleep(200 * time.Microsecond)
	}

}

func metricWorker(ID int, mChan chan []byte, udpClient net.Conn) {

	for {
		//log.Printf("Waiting message from chan")
		payload := <-mChan
		//time.Sleep(20 * time.Second)
		fmt.Fprintf(udpClient, string(payload))
		//udpClient.Write(payload)

	}

}

func mqttWorker(ID int, mChan chan []byte, client mqtt.Client) {
	log.Printf("starting mqtt worker %s", ID)
	topic := "sce/metrics"

	for {
		payload := <-mChan
		if token := client.Publish(topic, 0, false, string(payload)); token.Wait() && token.Error() != nil {
			log.Printf("error publishing [%v]", token.Error())
		}
	}
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

func newMqttClient() mqtt.Client {

	opts := mqtt.NewClientOptions().AddBroker("tcp://0.0.0.0:1883")

	client := mqtt.NewClient(opts)

RECONNECT:
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("error connection to broker [%v]", token.Error())
		time.Sleep(1 * time.Second)
		logrus.Errorf("Reconnecting ....")
		goto RECONNECT
	}

	return client
}

func main() {
	if err := agent.Listen(agent.Options{}); err != nil {
		log.Fatal(err)
	}

	clientA = newHTTPClient()
	clientB = newHTTPClient()

	mqttClient = newMqttClient()

	workers := 4

	metricChan = make(chan []byte, 4)

	for i := 0; i < workers; i++ {
		udpClient, _ = net.Dial("udp", "0.0.0.0:5140")
		go metricWorker(i, metricChan, udpClient)

		go mqttWorker(i, metricChan, mqttClient)
	}

	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}

	router := mux.NewRouter()
	router.HandleFunc("/", deafaultHandler)
	router.HandleFunc("/withoutgoroutine", withoutGoroutine)
	router.HandleFunc("/withgoroutine", withGoroutine)
	router.HandleFunc("/withsleepygoroutine", withSleepyGoroutine)
	router.HandleFunc("/withtelegraf", withTelegrafGoroutine)
	router.HandleFunc("/withtelegrafchan", withTelegrafToChannel)
	router.HandleFunc("/withmultitelegrafchan", withMultiTelegrafToChannel)
	router.HandleFunc("/withmultitelegrafjson", withMultiTelegrafJSON)
	router.HandleFunc("/withmqttchannel", withMqttChannel)
	router.Handle("/metrics", promhttp.Handler())

	listenAddress := fmt.Sprintf(":%s", port)
	log.Printf("Listening on %s", listenAddress)

	log.Fatal(http.ListenAndServe(listenAddress, router))
}
