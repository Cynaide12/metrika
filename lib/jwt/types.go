package lib_jwt

import "github.com/golang-jwt/jwt/v5"

var (
	AccessTokenJson  = "access_token"
	RefreshTokenJson = "refresh_token"
)

type JWTClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
