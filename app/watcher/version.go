package main

import "flag"

const Version = "0.1.0-dev"

var (
	printVersion bool
)

func init() {
	flag.BoolVar(&printVersion, "version", false, "print version")
}
