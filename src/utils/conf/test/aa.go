package main

import (
	"log"
	"utils/conf"
)

func Get2() {
	val, err := conf.Get("logs", "filename")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(val)
}

func main() {
	Get2()
}

func init() {
	err := conf.InitConf("./config/scm_config.ini")
	if err != nil {
		log.Fatal(err)
	}
}
