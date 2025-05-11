package main

import (
	// "encoding/json"
	"net/http"
	"time"
	"github.com/google/uuid"
	"workspace/github.com/Benjysparks/chirpy/internal/auth"
	"workspace/github.com/Benjysparks/chirpy/internal/database"
	"sort"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"` 
}

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp ID", err)
		return
	}

	dbChirp, err := cfg.db.GetChirpsByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		UserID:    dbChirp.UserID,
		Body:      dbChirp.Body,
	})
}

func (cfg *apiConfig) handlerChirpsRetrieve(w http.ResponseWriter, r *http.Request) {
	
	sortType := r.URL.Query().Get("sort")
	if sortType != "desc" {
		sortType = "asc"
	}

	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}


	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
				ID:        dbChirp.ID,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.UpdatedAt,
				UserID:    dbChirp.UserID,
				Body:      dbChirp.Body,
		})
	}

	finalChirps := []Chirp{}
	s := r.URL.Query().Get("author_id")
	if s == "" {
		finalChirps = chirps
	} else {

		authorString, err := uuid.Parse(s)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Could not parse", err)
			return
		}


		for _, dbChirp := range chirps {
			if dbChirp.UserID == authorString {
				finalChirps = append(finalChirps, dbChirp)
			}
		}
	}

	if sortType == "desc" {
		sort.Slice(finalChirps, func(i, j int) bool {
			return finalChirps[i].CreatedAt.After(finalChirps[j].CreatedAt)
		})
	} else {
		sort.Slice(finalChirps, func(i, j int) bool {
			return finalChirps[i].CreatedAt.Before(finalChirps[j].CreatedAt)
		})
	}
	



	respondWithJSON(w, http.StatusOK, finalChirps)
	
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError,"Could not parse id", err)
		return
	}
	
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No token found in header", err)
		return
	}

	authUser, err := auth.ValidateJWT(token, cfg.JwtSecret) 
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Could not validate token", err)
		return
	}

	chirp, err := cfg.db.GetChirpsByID(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Could not find chirp", err)
		return
	}

	if chirp.UserID != authUser {
		respondWithError(w, http.StatusForbidden, "not authorised to complete delete task", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), database.DeleteChirpParams{
		ID:         chirpID,
		UserID: 	authUser,
	})
	
	if err != nil {
		respondWithError(w, http.StatusForbidden, "Could not find chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

