package routes

import (
	"context"
	"errors"
	"fmt"
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
		AleardyRegister bool   `json:"AleardyRegister" example:"true" doc:"First register."`
	}
}

type UserOutput struct {
	Body struct {
		Users []model.User `json:"Users"`
	}
}

func User(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getuser",
		Method:      "GET",
		Path:        "/user",
	}, func(ctx context.Context, input *struct {
		User string `header:"user" example:"user" doc:"Username."`
	}) (*UserOutput, error) {
		if input.User == "" {
			resp := []*model.User{}
			output := &UserOutput{}
			res := global.DBEngine.Model(&model.User{}).Find(&resp)
			if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return nil, res.Error
			}
			for _, user := range resp {
				user.Password = ""
				output.Body.Users = append(output.Body.Users, *user)
			}
			return output, nil

		} else {
			resp := &model.User{}
			resp.Password = ""
			res := global.DBEngine.Model(&model.User{}).Where("username = ?", input.User).First(&resp)
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return nil, res.Error
			}
			output := &UserOutput{}
			output.Body.Users = append(output.Body.Users, *resp)
			return output, nil
		}
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
		fmt.Print(res.Error)
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			resp.Body.AleardyRegister = false
			social := &model.Social{}
			global.DBEngine.Create(social)
			user := &model.User{
				Username:      input.Username,
				Email:         input.Email,
				Password:      input.Password,
				JoinedTime:    global.DBEngine.NowFunc(),
				LasttimeLogin: global.DBEngine.NowFunc(),
				SocialID:      social.ID,
			}
			result := global.DBEngine.Create(user)
			if result.Error != nil {
				fmt.Println("Failed to create user:", result.Error)
				return nil, result.Error
			}
		} else {
			resp.Body.AleardyRegister = true
			fmt.Print(input)
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
		Path:        "/user",
		Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		Username    string `header:"username" example:"user" doc:"Username." required:"true"`
		Email       string `header:"email" example:"wow@gmail" doc:"Email."`
		Description string `header:"Description" example:"hello how are you" doc:"user description."`
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		fmt.Println(ctx.Value("userid"))
		result := global.DBEngine.Model(&model.User{}).Where("ID = ?", ctx.Value("userid")).Updates(map[string]interface{}{
			"Username":    input.Username,
			"Email":       input.Email,
			"Description": input.Description,
		})
		if result.Error != nil {
			fmt.Println("Failed to update user:", result.Error)
			resp.Body.Message = "Edit failed!"
			return resp, result.Error
		}
		resp.Body.Message = "Edit success!"

		return resp, nil
	})
}
