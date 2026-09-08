package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/virer/konsumo/web"
)

func TestMainRoutes(t *testing.T) {
	if err := os.Chdir("."); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll("data", os.ModePerm)

	mux := http.NewServeMux()
	mux.HandleFunc("/", web.HomeHandler)
	mux.HandleFunc("/submit", web.SubmitHandler)
	mux.HandleFunc("/delete", web.DeleteHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("ui/assets"))))

	server := httptest.NewServer(mux)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("GET / failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET / status = %d, want 200", resp.StatusCode)
	}
}
