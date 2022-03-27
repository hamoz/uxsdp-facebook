package main

import (
	"log"
	"net/http"

	"github.com/hamoz/uxsdp-facebook/fb"
)

func main() {
	http.HandleFunc("/webhook", fb.HandleMessenger)

	port := ":8099"
	log.Fatal(http.ListenAndServe(port, nil))
}
