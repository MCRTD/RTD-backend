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

type RegisterOutput struct {
	Body struct {
		Message string `json:"message" example:"Success" doc:"Status message."`
	}
}

func User(api huma.API) {

	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      "GET",
		Path:        "/user/register",
		// Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct{}) (*RegisterOutput, error) {
		resp := &RegisterOutput{}
		resp.Body.Message = "Register success!"

		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      "GET",
		Path:        "/user/login",
		// Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct{}) (*LoginOutput, error) {
		resp := &LoginOutput{}
		token, err := middleware.CreateToken("username")
		if err != nil {
			return nil, err
		}
		resp.Body.Message = "login success!"
		resp.SetCookie = append(resp.SetCookie, http.Cookie{Name: "token", Value: token.Token})
		resp.SetCookie = append(resp.SetCookie, http.Cookie{Name: "reflashtoken", Value: token.ReflashToken})

		return resp, nil
	})
}
