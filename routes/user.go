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
		Method:      "POST",
		Path:        "/user",
	}, func(ctx context.Context, input *struct {
		User string `header:"user" example:"user" doc:"Username."`
	}) (*model.User, error) {
		resp := &model.User{}
		resp.Password = ""
		res := global.DBEngine.Model(&model.User{}).Where("username = ?", input.User).First(&resp)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, res.Error
		}
		return resp, nil
	})

	// register
	huma.Register(api, huma.Operation{
		OperationID: "register",
		Method:      "POST",
		Path:        "/user/register",
		// Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		Username string `header:"username" example:"user" doc:"Username."`
		Password string `header:"password" example:"password" doc:"Password."`
		Email    string `header:"email" example:"wow@mail.com" doc:"Email."`
	}) (*RegisterOutput, error) {
		resp := &RegisterOutput{}
		resp.Body.Message = "Register success!"

		res := global.DBEngine.Model(&model.User{}).Where("username = ?", input.Username).First(&model.User{})
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			resp.Body.AleardyRegister = true
		} else {
			resp.Body.AleardyRegister = false
			global.DBEngine.Create(&model.User{
				Username:      input.Username,
				Email:         input.Email,
				Password:      input.Password,
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
		Username string `header:"username" example:"user" doc:"Username."`
		Password string `header:"password" example:"password" doc:"Password."`
	}) (*LoginOutput, error) {
		resp := &LoginOutput{}

		var user model.User

		if global.DBEngine.Model(&model.User{}).Where("username = ? AND password = ?", input.Username, input.Password).First(&user).Error != nil {
			return nil, errors.New("login failed")
		}
		token, err := middleware.CreateToken(user.ID)
		if err != nil {
			return nil, err
		}
		resp.Body.Message = "login success!"
		resp.SetCookie = append(resp.SetCookie, http.Cookie{Name: "token", Value: token.Token})
		resp.SetCookie = append(resp.SetCookie, http.Cookie{Name: "reflashtoken", Value: token.ReflashToken})

		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "edituser",
		Method:      "PATCH",
		Path:        "/user/edit",
		Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		Username    string `header:"username" example:"user" doc:"Username."`
		Email       string `header:"email" example:"wow@gmail" doc:"Email."`
		Description string `header:"Description" example:"hello how are you" doc:"user description."`
	}) (*RegisterOutput, error) {
		resp := &RegisterOutput{}
		resp.Body.Message = "Edit success!"
		global.DBEngine.Model(&model.User{}).Where("ID = ?", ctx.Value("userid")).Updates(&model.User{
			Username:    input.Username,
			Email:       input.Email,
			Description: input.Description,
		})

		return resp, nil
	})
}
