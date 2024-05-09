package routes

import (
	"context"
	"errors"
	"net/http"

	"RTD-backend/global"
	"RTD-backend/middleware"
	"RTD-backend/model"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

type LoginOutput struct {
	Body struct {
		Message string `json:"message" example:"Success" doc:"Status message."`
	}
	SetCookie []http.Cookie `header:"Set-Cookie"`
}

type RegisterOutput struct {
	Body struct {
		Message         string `json:"message" example:"Success" doc:"Status message."`
		AleardyRegister bool   `json:"firstRegister" example:"true" doc:"First register."`
	}
}

func User(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getuser",
		Method:      "GET",
		Path:        "/user",
	}, func(ctx context.Context, input *struct {
		user string `header:"user" example:"user" doc:"Username."`
	}) (*model.User, error) {
		resp := &model.User{}
		resp.Password = ""
		res := global.DBEngine.Model(&model.User{}).Where("username = ?", input.user).First(&resp)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, res.Error
		}
		return resp, nil
	})

	// register
	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      "GET",
		Path:        "/user/register",
		// Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		username string `header:"username" example:"user" doc:"Username."`
		password string `header:"password" example:"password" doc:"Password."`
		email    string `header:"email" example:"wow@mail.com" doc:"Email."`
	}) (*RegisterOutput, error) {
		resp := &RegisterOutput{}
		resp.Body.Message = "Register success!"

		res := global.DBEngine.Model(&model.User{}).Where("username = ?", input.username).First(&model.User{})
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			resp.Body.AleardyRegister = true
		} else {
			resp.Body.AleardyRegister = false
			global.DBEngine.Create(&model.User{
				Username:      input.username,
				Email:         input.email,
				Password:      input.password,
				JoinedTime:    global.DBEngine.NowFunc(),
				LasttimeLogin: global.DBEngine.NowFunc(),
			})
		}
		return resp, nil

	})
	// login
	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      "GET",
		Path:        "/user/login",
		// Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		username string `header:"username" example:"user" doc:"Username."`
		password string `header:"password" example:"password" doc:"Password."`
	}) (*LoginOutput, error) {
		resp := &LoginOutput{}
		token, err := middleware.CreateToken(input.username)
		if err != nil {
			return nil, err
		}
		if global.DBEngine.Model(&model.User{}).Where("username = ? AND password = ?", input.username, input.password).First(&model.User{}).Error != nil {
			return nil, errors.New("login failed")
		}

		resp.Body.Message = "login success!"
		resp.SetCookie = append(resp.SetCookie, http.Cookie{Name: "token", Value: token.Token})
		resp.SetCookie = append(resp.SetCookie, http.Cookie{Name: "reflashtoken", Value: token.ReflashToken})

		return resp, nil
	})
}
