package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secret-key")

func createToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Minute * 30).Unix(),
			"iat":      time.Now().Unix(),
		})

	reflashToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().AddDate(0, 1, 0).Unix(), // 1 month
			"iat":      time.Now().Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	reflashTokenString, err := reflashToken.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return `{"token":"` + tokenString + `","reflashToken":"` + reflashTokenString + `"}`, nil
}

func reflashToken(reflashTokenString string) (string, error) {
	token, err := jwt.Parse(reflashTokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", err
	}

	username := token.Claims.(jwt.MapClaims)["username"].(string)

	token = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Minute * 30).Unix(),
			"iat":      time.Now().Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func parseToken(api huma.API) func(ctx huma.Context, next func(huma.Context)) {

	return func(ctx huma.Context, next func(huma.Context)) {

		tokenString := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")

		if tokenString == "" {
			huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden")
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden")
			return
		}

		if !token.Valid {
			huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden")
			return
		}

		next(ctx)
	}
}
