package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/virer/konsumo/web"
)

func main() {
	addr := flag.String("addr", ":8080", "IP address and port to listen on (e.g., :8080, 0.0.0.0:8080, localhost:3000)")
	flag.Parse()

	os.MkdirAll("data", os.ModePerm)

	http.HandleFunc("/", web.HomeHandler)
	http.HandleFunc("/submit", web.SubmitHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("ui/assets"))))

	log.Printf("Running on http://%s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
