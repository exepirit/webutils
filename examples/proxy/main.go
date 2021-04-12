package main

import (
	"github.com/exepirit/webutils"
	"log"
	"net/http"
)

func HandleIndex(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Hello, World"))
}

func main() {
	proxy := webutils.NewProxy("https://example.com", "/home/")

	http.HandleFunc("/", HandleIndex)
	http.Handle("/home", proxy)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln(err)
	}
}
