package conf

import (
	"testing"
)

func TestGet(t *testing.T) {
	val, err := Get("local", "proxy_addr")
	if err != nil {
		t.Fatal(err)
	}
	if val != "0.0.0.0" {
		t.Fatal("val != local.proxy_addr")
	}
}

func TestGetSect(t *testing.T) {
	val, err := GetSect("local")
	if err != nil {
		t.Fatal(err)
	}
	if val["proxy_addr"] != "0.0.0.0" {
		t.Fatal("val != local.proxy_addr")
	}
}

func init() {
	InitConf("./config/scm_config.ini")
}
