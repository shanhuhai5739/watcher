package heartbeat

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

var (
	Version = "0.1"
)

func TestHeartbeat(t *testing.T) {
	url := "http://127.0.0.1:9091/heartbeat"
	h := &Heartbeat{
		Version:   Version,
		Hostname:  "localhost",
		Timestamp: time.Now().Unix(),
	}
	err := h.Callback(url)
	if err != nil {
		log.Fatal("Heartbeat: encode is err, err:%v", err)
	}
}

func handleHearbeat(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	//println("body: ", string(body))

	h, err := Decode(string(body))
	if err != nil {
		log.Fatal(err)
	}
	if h.Version != Version {
		log.Fatal("h.Version != Version")
	}
}

func init() {
	go func() {
		http.HandleFunc("/heartbeat", handleHearbeat)
		err := http.ListenAndServe(":9091", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()
}
