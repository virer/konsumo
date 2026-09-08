package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/virer/konsumo/web"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "IP address and port to listen on (e.g., 127.0.0.1:8080, 0.0.0.0:8080, :3000)")
	flag.Parse()

	os.MkdirAll("data", os.ModePerm)

	http.HandleFunc("/", web.HomeHandler)
	http.HandleFunc("/submit", web.SubmitHandler)
	http.HandleFunc("/delete", web.DeleteHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("ui/assets"))))

	log.Printf("Running on http://%s", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
