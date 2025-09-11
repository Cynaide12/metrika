package lib_jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var(
	AccessTokenTTL = time.Minute * 15
	RefreshTokenTTL = time.Hour * 24 * 30
	RefreshTokenMaxAge = int(RefreshTokenTTL.Seconds())
)

func GenerateAccessJWT(jwt_secret string, email string, userID uint, userRole string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:    email,
		Role:	userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwt_secret))
}

func GenerateRefreshJWT(jwt_secret string, email string, userID uint, userRole string) (string, error) {
	claims := JWTClaims{
		UserID: userID,
		Email:    email,
		Role: userRole,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwt_secret))
}

func GeneratePairJWT(jwt_secret string, email string, userID uint, userRole string) (accessToken string, refreshToken string, err error) {
	access, err := GenerateAccessJWT(jwt_secret, email, userID, userRole)
	if err != nil {
		return "", "", err
	}
	refresh, err := GenerateRefreshJWT(jwt_secret, email, userID, userRole)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}
