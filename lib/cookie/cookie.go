package lib_cookie

import (
	lib_jwt "metrika/lib/jwt"
	"net/http"
)

func AddCookie(w http.ResponseWriter, refreshToken string) http.ResponseWriter {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   lib_jwt.RefreshTokenMaxAge,
		SameSite: http.SameSiteLaxMode,
	})

	return w
}

func RemoveCookie(w http.ResponseWriter) http.ResponseWriter {
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
