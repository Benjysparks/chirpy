package auth

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	"github.com/google/uuid"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"fmt"
	"errors"
	"strings"
	"net/http"
	"crypto/rand"
	"encoding/hex"
)

func HashPassword(password string) (string, error) {
	pwordByte := []byte(password)
	hash, err := bcrypt.GenerateFromPassword(pwordByte, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	hashString := string(hash)
	return hashString, nil
}

func CheckPasswordHash(hash, password string) error {
	log.Print(hash)
	log.Print(password)
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	stringID := fmt.Sprintf("%v", userID)
	claims := &jwt.RegisteredClaims{
		ExpiresAt: 		jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Issuer:			"chirpy",
		IssuedAt:		jwt.NewNumericDate(time.Now().UTC()),
		Subject:		stringID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString([]byte(tokenSecret))
	return ss, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
    parser := jwt.NewParser(jwt.WithLeeway(5*time.Second))
    token, err := parser.ParseWithClaims(
        tokenString, 
        &jwt.RegisteredClaims{}, 
        func(token *jwt.Token) (interface{}, error) {
            return []byte(tokenSecret), nil
        },
    )
    
    if err != nil {
        return uuid.UUID{}, err
    }

    claims, ok := token.Claims.(*jwt.RegisteredClaims)
    if !ok {
        return uuid.UUID{}, errors.New("invalid claims type")
    }

    idStr := claims.Subject
    parsedID, err := uuid.Parse(idStr)
    if err != nil {
        return uuid.UUID{}, err
    }
    
    return parsedID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header not found")
	}
	
	// Check if it starts with "Bearer "
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("Authorization header format must be 'Bearer {token}'")
	}
	
	// Extract the token by removing "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)
	
	if token == "" {
		return "", errors.New("Token not found in Authorization header")
	}
	
	return token, nil
}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	rand.Read(key)
	encodedStr := hex.EncodeToString(key)
	return encodedStr, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header not found")
	}
	
	// Check if it starts with "Bearer "
	if !strings.HasPrefix(authHeader, "ApiKey ") {
		return "", errors.New("Authorization header format must be 'ApiKey {token}'")
	}
	
	// Extract the token by removing "Bearer " prefix
	token := strings.TrimPrefix(authHeader, "ApiKey ")
	token = strings.TrimSpace(token)
	
	if token == "" {
		return "", errors.New("Token not found in Authorization header")
	}
	
	return token, nil

}