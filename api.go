package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/debobrad579/chirpy/internal/database"
	"github.com/google/uuid"
)

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": msg}); err != nil {
		log.Printf("Error encoding response: %s", err)
	}
}

func respondWithJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("Error encoding response: %s", err)
	}
}

func apiMux(cfg *apiConfig) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	mux.HandleFunc("POST /chirps", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Body   string    `json:"body"`
			UserId uuid.UUID `json:"user_id"`
		}

		var params parameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if len(params.Body) > 140 {
			respondWithError(w, http.StatusBadRequest, "Chirp is too long")
			return
		}

		profaneWords := []string{
			"kerfuffle",
			"sharbert",
			"fornax",
		}

		words := strings.Split(params.Body, " ")
		newWords := make([]string, len(words))

		for i, word := range words {
			if slices.Contains(profaneWords, strings.ToLower(word)) {
				newWords[i] = "****"
			} else {
				newWords[i] = word
			}
		}

		cleanedBody := strings.Join(newWords, " ")

		chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanedBody, UserID: params.UserId})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create chirp")
			return
		}

		respondWithJSON(w, http.StatusCreated, chirp)
	})

	mux.HandleFunc("GET /chirps", func(w http.ResponseWriter, r *http.Request) {
		chirps, err := cfg.db.GetChirps(r.Context())
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get chirps")
			return
		}

		respondWithJSON(w, http.StatusOK, chirps)
	})

	mux.HandleFunc("GET /chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
		chirpID, err := uuid.Parse(r.PathValue("chirpID"))
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "chirpID is not a uuid")
			return
		}

		chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusNotFound, "Chirp not found")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Failed to get chirp")
			return
		}

		respondWithJSON(w, http.StatusOK, chirp)
	})

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email string `json:"email"`
		}

		var params parameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		user, err := cfg.db.CreateUser(r.Context(), params.Email)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed creating user")
			return
		}

		respondWithJSON(w, http.StatusCreated, user)
	})

	return mux
}
