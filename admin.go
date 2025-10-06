package main

import (
	"fmt"
	"net/http"
)

func adminMux(cfg *apiConfig) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `
			<html>
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
			</html>
		`, cfg.fileserverHits.Load())
	})

	mux.HandleFunc("POST /reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		cfg.resetHits()

		if err := cfg.db.DeleteAllUsers(r.Context()); err != nil {
			http.Error(w, "Failed to delete all users", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	return mux
}
