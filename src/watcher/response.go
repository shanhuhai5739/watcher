package watcher

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

type Response struct {
	Action    string `json:"action"`
	Code      int    `json:"code"`
	Msg       string `json:"msg"`
	MD5       string `json:"md5"`
	BeforeCmd Cmd    `json:"beforeCmd"`
	AfterCmd  Cmd    `json:"afterCmd"`
}

type Cmd struct {
	Success bool   `json:"success"`
	Out     string `json:"out"`
	Err     error  `json:"err"`
	Msg     string `json:"msg"`
}

func (r *Response) Encode() ([]byte, error) {
	bt, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return bt, err
}

func Decode(context string) (*Response, error) {
	res := &Response{}
	err := json.Unmarshal([]byte(context), res)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (r *Response) Callback(url string) (err error) {
	if len(url) == 0 {
		err = fmt.Errorf("Response: callback url is null")
		return
	}

	if r == nil {
		err = fmt.Errorf("Response: response is nil")
		return
	}

	jsonByte, err := r.Encode()
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
		err = fmt.Errorf("Response: callback to %v is err, err: %v", url, resp.Status)
		return
	}

	return
}
