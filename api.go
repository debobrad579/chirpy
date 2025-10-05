package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
)

func apiMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})
	mux.HandleFunc("POST /validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		type parameters struct {
			Body string `json:"body"`
		}

		type returnVals struct {
			Valid       bool   `json:"valid"`
			CleanedBody string `json:"cleaned_body,omitempty"`
			Error       string `json:"error,omitempty"`
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

		respondWithJSON(w, http.StatusOK, returnVals{Valid: true, CleanedBody: strings.Join(newWords, " ")})
	})

	return mux
}
