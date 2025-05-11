package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
	"github.com/google/uuid"
	"workspace/github.com/Benjysparks/chirpy/internal/database"
	"workspace/github.com/Benjysparks/chirpy/internal/auth"
	"log"
)

func (cfg *apiConfig) handlerChirpsValidate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body     string  `json:"body"`
		UserID  uuid.UUID  `json:"user_id"`
	}

	type Chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"` 
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid JWT token", err)
		return
	}

	JwtUser, err := auth.ValidateJWT(token, cfg.JwtSecret) 
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "JWT token not valid", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		log.Printf("")
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	splitChirp := strings.Split(params.Body, " ")

	for index, word := range splitChirp {
		temp := strings.ToLower(word)
		if temp == "kerfuffle" || temp == "sharbert" || temp == "fornax" {
			splitChirp[index] = "****"
		}
	}

	cleanedChirp := strings.Join(splitChirp, " ")


	chirp, err := cfg.db.NewChirp(r.Context(), database.NewChirpParams{
		Body: cleanedChirp,
		UserID: JwtUser,
	})
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Can not post chirp!", err)  // Fix: proper error handling
        return
    }
    
    respondWithJSON(w, http.StatusCreated, Chirp{  // Fix: use http.StatusCreated (201)
        ID:        chirp.ID,
        CreatedAt: chirp.CreatedAt,
        UpdatedAt: chirp.UpdatedAt,
        Body:      chirp.Body,
		UserID:    JwtUser,
    })
}
