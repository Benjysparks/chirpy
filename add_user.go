package main

import (
	"net/http"
	"encoding/json"
	"time"
	"github.com/google/uuid"
    "workspace/github.com/Benjysparks/chirpy/internal/auth"
    "workspace/github.com/Benjysparks/chirpy/internal/database"
    "database/sql"
    "fmt"
)

type User struct {
    ID              uuid.UUID `json:"id"`
    Username        string    `json:"username"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
    Email           string    `json:"email"`
    Token           string    `json:"token"`
    RefreshToken    string    `json:"refresh_token"`
    IsChirpyRed     bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerAddUser(w http.ResponseWriter, r *http.Request) {
    type parameters struct {
        Email string `json:"email"`
        Password string `json:"password"`
        Username string  `json:"username"`
    }

    type User struct {
        ID              uuid.UUID `json:"id"`
        Username        string    `json:"username"`
        CreatedAt       time.Time `json:"created_at"`
        UpdatedAt       time.Time `json:"updated_at"`
        Email           string    `json:"email"`
        HashedPassword  string    `json:"hashed_password"` 
        Token           string    `json:"token"`
        RefreshToken    string    `json:"refresh_token"`
        IsChirpyRed     bool      `json:"is_chirpy_red"`
    }
    
    decoder := json.NewDecoder(r.Body)  // Fix: use r.Body instead of r.Email
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
        return
    }

    hashedPassword, err := auth.HashPassword(params.Password)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
    }
    
    

    dbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
        Email:              params.Email,
        HashedPassword:     sql.NullString{String: hashedPassword, Valid: true},
        Username:           params.Username,
    })
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)  // Fix: proper error handling
        return
    }
    
    respondWithJSON(w, http.StatusCreated, User{  // Fix: use http.StatusCreated (201)
        ID:        dbUser.ID,
        Username:  dbUser.Username,
        CreatedAt: dbUser.CreatedAt,
        UpdatedAt: dbUser.UpdatedAt,
        Email:     dbUser.Email,
        HashedPassword:  hashedPassword,
        IsChirpyRed: dbUser.IsChirpyRed.Bool,
    })
}


func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {

    type parameters struct {
        Email string `json:"email"`
        Password string `json:"password"`
        ExpiresInSeconds time.Duration `json:"expires_in_seconds,omitempty"`
    }   

    type User struct {
		ID              uuid.UUID `json:"id"`
        Username        string    `json:"username"`
		CreatedAt       time.Time `json:"created_at"`
		UpdatedAt       time.Time `json:"updated_at"`
		Email           string    `json:"email"`
        Token           string    `json:"token"`
        RefreshToken    string    `json:"refresh_token"`
        IsChirpyRed     bool      `json:"is_chirpy_red"`
	}

    decoder := json.NewDecoder(r.Body)  // Fix: use r.Body instead of r.Email
    params := parameters{}
    err := decoder.Decode(&params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
        return
    }

    if params.ExpiresInSeconds == 0 {
        params.ExpiresInSeconds = 3600  // Default to 1 hour
    }
    if params.ExpiresInSeconds > 3600 {
        params.ExpiresInSeconds = 3600  // Cap at 1 hour
    }
    // Convert to duration
    expiryDuration := time.Duration(params.ExpiresInSeconds) * time.Second
    

    user, err := cfg.db.SearchEmail(r.Context(), params.Email)
    if err != nil {
        // If user not found or other database error
        respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
        return
    }
    

    hashedString := fmt.Sprintf("%v", user.HashedPassword.String)
    err = auth.CheckPasswordHash(hashedString, params.Password)
    if err != nil {
        // Password doesn't match
        respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
        return
    }

    userToken, err := auth.MakeJWT(user.ID, cfg.JwtSecret, expiryDuration)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Could not create token", nil)
    }

    rToken, _ := auth.MakeRefreshToken()

    dbrToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
        Token:     rToken,
	    UserID:    user.ID,
	    ExpiresAt: time.Now().AddDate(0, 0, 60),
    })

    // If we get here, authentication succeeded
    respondWithJSON(w, http.StatusOK, User{
        ID:             user.ID,
        Username:       user.Username,
        CreatedAt:      user.CreatedAt,
        UpdatedAt:      user.UpdatedAt,
        Email:          user.Email,
        Token:          userToken, 
        RefreshToken:   dbrToken.Token,
        IsChirpyRed:    user.IsChirpyRed.Bool,     
    })
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
    
    type User struct {
        ID              uuid.UUID `json:"id"`
        Username        string    `json:"username"`
        CreatedAt       time.Time `json:"created_at"`
        UpdatedAt       time.Time `json:"updated_at"`
        Email           string    `json:"email"`
        Token           string    `json:"token"`
        RefreshToken    string    `json:"refresh_token"`
        IsChirpyRed     bool      `json:"is_chirpy_red"`
    }
    
    refreshToken, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "No refresh token found in header", err)
        return 
    }

    checkedRToken, err := cfg.db.GetRefreshToken(r.Context(), refreshToken)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Could not find refresh token", err)
        return
    }

    if checkedRToken.ExpiresAt.Compare(time.Now()) == -1 {
        respondWithError(w, http.StatusUnauthorized, "refresh token expired", nil)
        return
    }


    if checkedRToken.RevokedAt.Valid {
        respondWithError(w, http.StatusUnauthorized, "Refresh token has been revoked", err)
        return
    }

    refreshedUser, err := cfg.db.GetUserFromRToken(r.Context(), checkedRToken.Token)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Could not find user in database", err)
        return
    }

    hour, _ := time.ParseDuration("1h")

    userToken, err := auth.MakeJWT(refreshedUser.UserID, cfg.JwtSecret, hour)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Could not create token", nil)
    }

    respondWithJSON(w, http.StatusOK, User{
        Token:  userToken,
    })
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {

    refreshToken, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "No refresh token found in header", err)
        return 
    }

    err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "No refresh token found in database", err)
        return 
    }

    w.WriteHeader(http.StatusNoContent)

}

func (cfg *apiConfig) handlerChangePassword(w http.ResponseWriter, r *http.Request) {

    type UserInfo struct {
        Email       string  `json:"email"`
        Password    string  `json:"password"`  
    }

    accessToken, err := auth.GetBearerToken(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "No token found in header", err)
        return 
    }

    UserToUpdate, err := auth.ValidateJWT(accessToken, cfg.JwtSecret)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Could not validate token", err)
        return
    }

    decoder := json.NewDecoder(r.Body)  // Fix: use r.Body instead of r.Email
    newInfo := UserInfo{}
    err = decoder.Decode(&newInfo)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Couldn't decode parameters", err)
        return
    }

    hashedPassword, err := auth.HashPassword(newInfo.Password)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Could not hash password", err)
        return
    }

    err = cfg.db.UpdateUserInfo(r.Context(), database.UpdateUserInfoParams{
        Email:              newInfo.Email,
	    HashedPassword:     sql.NullString{String: hashedPassword, Valid: true},
	    ID:                 UserToUpdate,
    })
    if err != nil {
        respondWithError(w, http.StatusUnauthorized, "Could not update user info", err)
        return
    }

    respondWithJSON(w, http.StatusOK, UserInfo{
        Email: newInfo.Email,
    })

}

func (cfg *apiConfig) handlerUpgradeAccount(w http.ResponseWriter, r *http.Request) {

    type parameters struct {
        Event   string      `json:"event"`
        Data struct {
            UserID string `json:"user_id"`
        } `json:"data"`
        }

    apiKey, err := auth.GetAPIKey(r.Header)
    if err != nil {
        respondWithError(w, http.StatusUnauthorized,"No key found in header", err)
        return
    }

    if apiKey != cfg.polkaKey {
        respondWithError(w, http.StatusUnauthorized,"Incorrect key found in header", err)
        return
    }

    decoder := json.NewDecoder(r.Body)  // Fix: use r.Body instead of r.Email
    params := parameters{}
    err = decoder.Decode(&params)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
        return
    }
    
    if params.Event != "user.upgraded" {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    parsedUser, err := uuid.Parse(params.Data.UserID)
    if err != nil {
		respondWithError(w, http.StatusInternalServerError,"Could not parse id", err)
		return
	}

    err = cfg.db.UpgradeUser(r.Context(), parsedUser)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Could not find user", err)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerShowUsers(w http.ResponseWriter, r *http.Request) {
    users, err := cfg.db.GetAllUsers(r.Context())
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Could not find users", err)
        return
    }

    allUsers := []User{}

    for _, user := range users {
        allUsers = append(allUsers, User{
            ID:     user.ID,
            Username: user.Username,
            Email:  user.Email,
        })
    }

    respondWithJSON(w, http.StatusOK, allUsers)
}