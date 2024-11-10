package routes

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"RTD-backend/global"
	"RTD-backend/middleware"
	"RTD-backend/model"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	storage_go "github.com/supabase-community/storage-go"
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

type EditUser struct {
	Username    string `example:"user" doc:"Username."`
	Email       string `example:"wow@gmail" doc:"Email."`
	Description string `example:"hello how are you" doc:"user description."`
}

type EditUserPassword struct {
	Password string `example:"password" doc:"Password."`
}

func User(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getuser",
		Method:      "GET",
		Path:        "/user",
	}, func(ctx context.Context, input *struct {
		User string `query:"user" example:"user" doc:"Username."`
	}) (*UserOutput, error) {
		if input.User == "" {
			resp := []*model.User{}
			output := &UserOutput{}
			res := global.DBEngine.Preload("Litematicas").Preload("Groups").Preload("Servers").Preload("Social").Model(&model.User{}).Find(&resp)
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
			res := global.DBEngine.Preload("Litematicas").Preload("Groups").Preload("Servers").Preload("Social").Model(&model.User{}).Where("ID = ?", input.User).First(&resp)
			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return nil, res.Error
			}
			resp.Password = ""
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
		Body EditUser
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		result := global.DBEngine.Model(&model.User{}).Where("ID = ?", ctx.Value("userid")).Updates(map[string]interface{}{
			"Username":    input.Body.Username,
			"Email":       input.Body.Email,
			"Description": input.Body.Description,
		})
		if result.Error != nil {
			fmt.Println("Failed to update user:", result.Error)
			resp.Body.Message = "Edit failed!"
			return resp, result.Error
		}
		resp.Body.Message = "Edit success!"

		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "edituserpassword",
		Method:      "PATCH",
		Path:        "/user/password",
		Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		Body EditUserPassword
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		result := global.DBEngine.Model(&model.User{}).Where("ID = ?", ctx.Value("userid")).Updates(map[string]interface{}{
			"Password": input.Body.Password,
		})
		if result.Error != nil {
			fmt.Println("Failed to update user password:", result.Error)
			resp.Body.Message = "Edit failed!"
			return resp, result.Error
		}
		resp.Body.Message = "Edit success!"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "adduseravatar",
		Method:      "POST",
		Path:        "/user/avatar",
		Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		RawBody multipart.Form
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		user := &model.User{}
		global.DBEngine.Model(&model.User{}).Where("ID = ?", ctx.Value("userid")).First(&user)
		if user.Avatar != "" {
			global.S3Client.RemoveFile("images", []string{user.AvatarPath})
		}
		file := input.RawBody.File["avatar"][0]
		filedata, err := file.Open()
		if err != nil {
			resp.Body.Message = "Failed"
			return resp, huma.Error400BadRequest("Failed to open file")
		}
		if file.Size > 1024*1024*10 {
			resp.Body.Message = "Failed"
			return resp, huma.NewError(413, "File is too large")
		}
		parts := strings.Split(file.Filename, ".")
		newfilename := uuid.New().String() + "." + parts[len(parts)-1]
		// 讀取文件的前512個字節來檢測內容類型
		buffer := make([]byte, 512)
		_, err = filedata.Read(buffer)
		if err != nil {
			fmt.Println("Failed to read file:", err)
			resp.Body.Message = "Failed"
			return resp, err
		}
		_, err = filedata.Seek(0, io.SeekStart)
		if err != nil {
			fmt.Println("Failed to seek file:", err)
			resp.Body.Message = "Failed"
			return resp, err
		}
		filetype := http.DetectContentType(buffer)
		global.S3Client.UploadFile("images", newfilename, filedata, storage_go.FileOptions{
			ContentType: &filetype,
		})
		url := global.S3Client.GetPublicUrl("images", newfilename)
		result := global.DBEngine.Model(&model.User{}).Where("ID = ?", ctx.Value("userid")).Updates(map[string]interface{}{
			"Avatar":     url.SignedURL,
			"AvatarPath": newfilename,
		})
		if result.Error != nil {
			fmt.Println("Failed to update user avatar:", result.Error)
			resp.Body.Message = "Edit failed!"
			return resp, result.Error
		}
		resp.Body.Message = "Edit success!"
		return resp, nil
	})

	// huma.Register(api, huma.Operation{
	// 	OperationID: "deleteuseravatar",
	// 	Method:      "DELETE",
	// 	Path:        "/user/avatar",
	// 	Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	// }, func(ctx context.Context, input *struct {
	// }) (*NormalOutput, error) {
	// 	resp := &NormalOutput{}
	// 	user := &model.User{}
	// 	result := global.DBEngine.Model(&model.User{}).Where("ID = ?", ctx.Value("userid")).First(&user)
	// 	if result.Error != nil {
	// 		fmt.Println("Failed to find user:", result.Error)
	// 		resp.Body.Message = "Delete failed!"
	// 		return resp, result.Error
	// 	}
	// 	if user.Avatar == "" {
	// 		resp.Body.Message = "Delete failed!"
	// 		return resp, huma.Error400BadRequest("User has no avatar")
	// 	}
	// 	global.S3Client.RemoveFile("images", []string{user.AvatarPath})
	// 	result = global.DBEngine.Model(&model.User{}).Where("ID = ?", ctx.Value("userid")).Updates(map[string]interface{}{
	// 		"Avatar":     "",
	// 		"AvatarPath": "",
	// 	})
	// 	if result.Error != nil {
	// 		fmt.Println("Failed to delete user avatar:", result.Error)
	// 		resp.Body.Message = "Delete failed!"
	// 		return resp, result.Error
	// 	}
	// 	resp.Body.Message = "Delete success!"
	// 	return resp, nil
	// })
}
