package routes

import (
	"context"
	"net/http"

	"RTD-backend/middleware"

	"github.com/danielgtaylor/huma/v2"
)

type LoginOutput struct {
	Body struct {
		Message string `json:"message" example:"Success" doc:"Status message."`
	}
	SetCookie []http.Cookie `header:"Set-Cookie"`
}

func Register(api huma.API) {

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      "GET",
		Path:        "/user/register",
		// Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct{}) (*LoginOutput, error) {
		resp := &LoginOutput{}
		token, err := middleware.CreateToken("username")
		if err != nil {
			return nil, err
		}
		resp.Body.Message = "Register success!"
		resp.SetCookie = append(resp.SetCookie, http.Cookie{Name: "token", Value: token.Token})
		resp.SetCookie = append(resp.SetCookie, http.Cookie{Name: "reflashtoken", Value: token.ReflashToken})

		return resp, nil
	})
}
