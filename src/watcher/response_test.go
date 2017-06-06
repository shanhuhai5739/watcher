package watcher

import (
	"testing"
)

var (
	str = `{"Code":200}`
)

func TestEncode(t *testing.T) {
	r := Response{Code: 200}
	ret, err := r.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if len(string(ret)) == 0 {
		t.Fatal("ret is null")
	}
}

func TestDecode(t *testing.T) {
	r := Response{Code: 200}
	ret, err := r.Encode()
	if err != nil {
		t.Fatal(err)
	}

	if len(string(ret)) == 0 {
		t.Fatal("ret is null")
	}

	response, err := Decode(string(ret))
	if err != nil {
		t.Fatal(err)
	}

	if response.Code != r.Code {
		t.Fatal("response code is err")
	}
}
