package main

import (
	"flag"
	"fmt"

	"github.com/pranavgore09/lkv/config"
	"github.com/pranavgore09/lkv/server"
)

func setupFlags() {
	flag.IntVar(&config.Port, "port", 7379, "Port for LKV")
	flag.StringVar(&config.Host, "host", "0.0.0.0", "Host LKV")
	flag.Parse()
}

func main() {
	setupFlags()
	fmt.Println("Starting LocalKeyValue Store Now...")
	server.AsyncRun()
}
