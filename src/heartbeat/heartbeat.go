package heartbeat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

var (
	respTimeout = 60 * time.Second
)

type Heartbeat struct {
	Version   string `json:"version"`
	Hostname  string `json:"hostname"`
	Timestamp int64  `json:"timestamp"`
	//Ip        string `json:"ip"`
	//LiveTime  string `json:"livetime"`
}

func (h *Heartbeat) Encode() ([]byte, error) {
	bt, err := json.Marshal(h)
	if err != nil {
		return nil, err
	}
	return bt, err
}

func Decode(context string) (*Heartbeat, error) {
	res := &Heartbeat{}
	err := json.Unmarshal([]byte(context), res)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (h *Heartbeat) Callback(url string) (err error) {
	if len(url) == 0 {
		err = fmt.Errorf("Heartbeat: url is null")
		return
	}

	if h == nil {
		err = fmt.Errorf("Heartbeat: heartbeat is nil")
		return
	}

	jsonByte, err := h.Encode()
	if err != nil {
		return
	}

	client := http.Client{Timeout: respTimeout}
	body := strings.NewReader(string(jsonByte))
	resp, err := client.Post(url, "application/json", body)
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("Heartbeat: post to %v is err, err: %v", url, resp.Status)
		return
	}

	return
}
