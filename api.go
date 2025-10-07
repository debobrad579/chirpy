package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/debobrad579/chirpy/internal/auth"
	"github.com/debobrad579/chirpy/internal/database"
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
			Body string `json:"body"`
		}

		userID, err := auth.AuthenticateUser(r.Header, cfg.tokenSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		var params parameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			respondWithError(w, http.StatusBadRequest, "Unauthorized")
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

		chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{Body: cleanedBody, UserID: userID})
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

	mux.HandleFunc("DELETE /chirps/{chirpID}", func(w http.ResponseWriter, r *http.Request) {
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

		userID, err := auth.AuthenticateUser(r.Header, cfg.tokenSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		if userID != chirp.UserID {
			respondWithError(w, http.StatusForbidden, "Forbidden")
			return
		}

		if err := cfg.db.DeleteChirp(r.Context(), chirp.ID); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed deleting chirp")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		var params parameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{Email: params.Email, HashedPassword: hashedPassword})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed creating user")
			return
		}

		respondWithJSON(w, http.StatusCreated, user)
	})

	mux.HandleFunc("PUT /users", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email    string
			Password string
		}

		userID, err := auth.AuthenticateUser(r.Header, cfg.tokenSecret)
		if err != nil {
			respondWithError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}

		var params parameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
		}

		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{ID: userID, Email: params.Email, HashedPassword: hashedPassword})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update user")
			return
		}

		respondWithJSON(w, http.StatusOK, user)
	})

	mux.HandleFunc("POST /login", func(w http.ResponseWriter, r *http.Request) {
		type parameters struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		type returnVals struct {
			ID           uuid.UUID `json:"id"`
			CreatedAt    time.Time `json:"created_at"`
			UpdatedAt    time.Time `json:"updated_at"`
			Email        string    `json:"email"`
			Token        string    `json:"token"`
			RefreshToken string    `json:"refresh_token"`
		}

		var params parameters
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusUnauthorized, "Email or password is incorrect")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Failed to get user")
			return
		}

		ok, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to check password hash")
			return
		}

		if !ok {
			respondWithError(w, http.StatusUnauthorized, "Email or password is incorrect")
			return
		}

		token, err := auth.MakeJWT(user.ID, cfg.tokenSecret, 1*time.Hour)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create access token")
			return
		}

		refreshTokenString, err := auth.MakeRefreshToken()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token")
			return
		}

		refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: refreshTokenString, UserID: user.ID, ExpiresAt: time.Now().Add(60 * 24 * time.Hour)})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create refresh token")
			return
		}

		respondWithJSON(w, http.StatusOK, returnVals{user.ID, user.CreatedAt, user.UpdatedAt, user.Email, token, refreshToken.Token})
	})

	mux.HandleFunc("POST /refresh", func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid authorization header")
			return
		}

		type returnVals struct {
			Token string `json:"token"`
		}

		refreshToken, err := cfg.db.GetRefreshToken(r.Context(), token)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Failed to get refresh token")
			return
		}

		if refreshToken.ExpiresAt.Before(time.Now()) || refreshToken.RevokedAt.Valid {
			respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
			return
		}

		accessToken, err := auth.MakeJWT(refreshToken.UserID, cfg.tokenSecret, 1*time.Hour)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create access token")
			return
		}

		respondWithJSON(w, http.StatusOK, returnVals{accessToken})
	})

	mux.HandleFunc("POST /revoke", func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid authorization header")
			return
		}

		if err := cfg.db.RevokeRefreshToken(r.Context(), token); err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusUnauthorized, "Invalid refresh token")
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Failed revoking refresh token")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	return mux
}
