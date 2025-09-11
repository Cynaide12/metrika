package lib_jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


func ValidateJWT(jwtSecret string, tokenString string) (*JWTClaims, error) {
	if tokenString == "" || jwtSecret == "" {
		return nil, fmt.Errorf("empty token or secret")
	}

	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}


	if claims.ExpiresAt != nil {
		if claims.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("token has expired")
		}
	} else {
		return nil, fmt.Errorf("token has no expiration")
	}


	if claims.Email == "" || claims.Role == "" || claims.UserID == 0{
		return nil, fmt.Errorf("token has no required claims")
	}

	return claims, nil
}
