package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/debobrad579/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             database.Queries
	platform       string
	tokenSecret    string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetHits() {
	cfg.fileserverHits.Store(0)
}

const (
	port         = "8080"
	filepathRoot = "."
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	tokenSecret := os.Getenv("TOKEN_SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to open database")
	}

	mux := http.NewServeMux()

	srv := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	cfg := &apiConfig{
		db:          *database.New(db),
		platform:    platform,
		tokenSecret: tokenSecret,
	}

	mux.Handle("/app/", http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir(filepathRoot)))))
	mux.Handle("/api/", http.StripPrefix("/api", apiMux(cfg)))
	mux.Handle("/admin/", http.StripPrefix("/admin", adminMux(cfg)))

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}
