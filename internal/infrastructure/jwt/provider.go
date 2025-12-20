package jwt

import (
	"fmt"
	domain "metrika/internal/domain/auth"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)


type JWTProvider struct {
	jwt_secret string
}

var (
	AccessTokenTTL     = time.Minute * 15
	RefreshTokenTTL    = time.Hour * 24 * 30
	RefreshTokenMaxAge = int(RefreshTokenTTL.Seconds())
	AccessTokenJson    = "access_token"
	RefreshTokenJson   = "refresh_token"
)

func NewJwtProvider(jwt_secret string) *JWTProvider {
	return &JWTProvider{jwt_secret}
}

func (p *JWTProvider) GenerateAccessJWT(email string, userID uint, session_id uint) (string, error) {
	claims := domain.JWTClaims{
		UserID:    userID,
		Email:     email,
		SessionID: session_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.jwt_secret))
}

func (p *JWTProvider) GenerateRefreshJWT(email string, userID uint, session_id uint) (string, error) {
	claims := domain.JWTClaims{
		UserID:    userID,
		Email:     email,
		SessionID: session_id,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.jwt_secret))
}

func (p *JWTProvider) GeneratePair(email string, userID uint, session_id uint) (tokens *domain.Tokens, err error) {
	access, err := p.GenerateAccessJWT(email, userID, session_id)
	if err != nil {
		return nil, err
	}
	refresh, err := p.GenerateRefreshJWT(email, userID, session_id)
	if err != nil {
		return nil, err
	}
	return &domain.Tokens{Access: access, Refresh: refresh}, nil
}

func (p JWTProvider) Validate(tokenString string) (*domain.JWTClaims, error) {
	if tokenString == "" || p.jwt_secret == "" {
		return nil, fmt.Errorf("empty token or secret")
	}

	claims := &domain.JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(p.jwt_secret), nil
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

	if claims.Email == "" || claims.SessionID == 0 || claims.UserID == 0 {
		return nil, fmt.Errorf("token has no required claims")
	}

	return claims, nil
}

func (p *JWTProvider) SetCookie(w http.ResponseWriter, refreshToken string) http.ResponseWriter {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   RefreshTokenMaxAge,
		SameSite: http.SameSiteLaxMode,
	})

	return w
}

func (p *JWTProvider) RemoveCookie(w http.ResponseWriter) http.ResponseWriter {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})

	return w
}
