package auth

import "github.com/golang-jwt/jwt/v5"

type Tokens struct {
	Access  string
	Refresh string
}

type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	SessionID uint   `json:"session_id"`
	jwt.RegisteredClaims
}
