package main

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	port         = "8080"
	filepathRoot = "."
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		log.Printf("Error encoding response: %s", err)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding response: %s", err)
	}
}

func main() {
	mux := http.NewServeMux()

	srv := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	cfg := &apiConfig{}

	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.Handle("/api/", http.StripPrefix("/api", apiMux()))
	mux.Handle("/admin/", http.StripPrefix("/admin", adminMux(cfg)))

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
