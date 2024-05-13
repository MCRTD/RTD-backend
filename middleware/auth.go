package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("secret-key")

type Auth struct {
	Token        string `json:"token"`
	ReflashToken string `json:"reflashToken"`
}

func CreateToken(UserID uint) (*Auth, error) {
	Token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userid": strconv.FormatUint(uint64(UserID), 10),
			"exp":    time.Now().Add(time.Minute * 30).Unix(),
			"iat":    time.Now().Unix(),
		})

	reflashToken := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userid": strconv.FormatUint(uint64(UserID), 10),
			"exp":    time.Now().AddDate(0, 1, 0).Unix(), // 1 month
			"iat":    time.Now().Unix(),
		})

	tokenString, err := Token.SignedString(secretKey)
	if err != nil {
		return nil, err
	}
	reflashTokenString, err := reflashToken.SignedString(secretKey)
	if err != nil {
		return nil, err
	}

	reruentoken := &Auth{
		Token:        tokenString,
		ReflashToken: reflashTokenString,
	}

	return reruentoken, nil
}

func ReflashToken(reflashTokenString string) (string, error) {
	token, err := jwt.Parse(reflashTokenString, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", err
	}

	userid := token.Claims.(jwt.MapClaims)["userid"].(string)

	token = jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userid": userid,
			"exp":    time.Now().Add(time.Minute * 30).Unix(),
			"iat":    time.Now().Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ReflashHandler(ctx huma.Context, next func(huma.Context)) {
	tokenString, err := huma.ReadCookie(ctx, "token")
	if err != nil || tokenString == nil {
		next(ctx)
		return
	}
	_, parseErr := jwt.Parse(tokenString.Value, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if parseErr == nil {
		next(ctx)
		return
	}
	if parseErr.Error() != "token has invalid claims: token is expired" {
		next(ctx)
		return
	}
	reflashTokenString, err := huma.ReadCookie(ctx, "reflashtoken")
	if err != nil {
		next(ctx)
		return
	}
	token, err := ReflashToken(reflashTokenString.Value)
	if err != nil {
		return
	}
	cookie := http.Cookie{
		Name:  "token",
		Path:  "/",
		Value: token,
	}
	fmt.Println("reflash token")
	ctx.AppendHeader("Set-Cookie", cookie.String())
	next(ctx)
}

func ParseToken(api huma.API) func(ctx huma.Context, next func(huma.Context)) {

	return func(ctx huma.Context, next func(huma.Context)) {

		// tokenString := strings.TrimPrefix(ctx.Header("Authorization"), "Bearer ")
		// from cookie get token

		tokenString, err := huma.ReadCookie(ctx, "token")
		if err != nil {
			huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden")
			return
		}

		if tokenString.Value == "" {
			huma.WriteErr(api, ctx, http.StatusForbidden, "Forbidden")
			return
		}

		token, err := jwt.Parse(tokenString.Value, func(token *jwt.Token) (interface{}, error) {
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

		ctx = huma.WithValue(ctx, "userid", token.Claims.(jwt.MapClaims)["userid"].(string))

		next(ctx)
	}
}
