package main

import (
	"log"
	"net/http"
	"os"

	"github.com/virer/konsumo/web"
)

func main() {
	os.MkdirAll("data", os.ModePerm)

	http.HandleFunc("/", web.HomeHandler)
	http.HandleFunc("/submit", web.SubmitHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	log.Println("Running on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
