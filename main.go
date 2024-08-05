package main

import (
	"cors/internal/middlewire/cors"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		w.Write([]byte("Hello World!"))
	})
	handle := cors.Default().Handler(mux)
	http.ListenAndServe(":8080", handle)
}
