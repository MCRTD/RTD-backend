package routes

import (
	"RTD-backend/global"
	"RTD-backend/middleware"
	"RTD-backend/model"
	"context"
	"fmt"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
)

type ServerOutput struct {
	Body struct {
		Servers []model.Server `json:"servers"`
	}
}

type ServerAdd struct {
	ServerName  string `json:"ServerName"`
	Description string `json:"Description"`
	Avatar      string `json:"Avatar"`
}

type ServerEdit struct {
	ServerID    uint   `json:"ServerID"`
	ServerName  string `json:"ServerName"`
	Description string `json:"Description"`
	Avatar      string `json:"Avatar"`
}

func Server(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getserver",
		Method:      "GET",
		Path:        "/server",
	}, func(ctx context.Context, input *struct {
		Serverid string `query:"serverid" example:"1" doc:"serverid."`
	}) (*ServerOutput, error) {
		resp := &ServerOutput{}
		var servers []model.Server
		if input.Serverid == "" {
			global.DBEngine.Model(&model.Server{}).Find(&servers)
		} else {
			global.DBEngine.Model(&model.Server{}).Where("ID = ?", input.Serverid).Find(&servers)
		}
		resp.Body.Servers = servers
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "add server",
		Method:      "POST",
		Path:        "/server",
		Middlewares: huma.Middlewares{middleware.IsAdmin(api)},
	}, func(ctx context.Context, input *struct {
		Body ServerAdd
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		social := &model.Social{}
		global.DBEngine.Create(social)
		server := &model.Server{
			ServerName:  input.Body.ServerName,
			Description: input.Body.Description,
			Avatar:      input.Body.Avatar,
			SocialID:    social.ID,
		}
		result := global.DBEngine.Create(server)
		if result.Error != nil {
			fmt.Println(result.Error)
			return nil, result.Error
		}
		resp.Body.Message = "success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "patch server",
		Method:      "PATCH",
		Path:        "/server",
		Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		Body ServerEdit
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		userIDStr := ctx.Value("userid").(string)
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			return nil, err
		}

		var server model.Server
		result := global.DBEngine.Preload("ServerOwner").Preload("ServerAdmins").First(&server, "id = ?", input.Body.ServerID)
		if result.Error != nil {
			return nil, fmt.Errorf("Server not found")
		}

		isAdmin := false
		//如果是伺服器擁有者或管理:p
		if server.ServerOwner.ID == uint(userID) {
			isAdmin = true
		} else {
			for _, admin := range server.ServerAdmins {
				if admin.ID == uint(userID) {
					isAdmin = true
					break
				}
			}
		}
		// 如果是管理員
		var user model.User
		result = global.DBEngine.First(&user, "id = ?", userID)
		if result.Error != nil {
			return nil, result.Error
		}
		if !isAdmin && !user.Admin {
			return nil, fmt.Errorf("Forbidden")
		}

		server.ServerName = input.Body.ServerName
		server.Description = input.Body.Description
		server.Avatar = input.Body.Avatar

		result = global.DBEngine.Save(&server)
		if result.Error != nil {
			return nil, result.Error
		}

		resp.Body.Message = "success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "delete server",
		Method:      "DELETE",
		Path:        "/server",
		Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		Serverid string `query:"serverid" example:"1" doc:"Server ID."`
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		userIDStr := ctx.Value("userid").(string)
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			return nil, err
		}
		var user model.User
		result := global.DBEngine.First(&user, "id = ?", userID)
		if result.Error != nil || !user.Admin {
			return nil, fmt.Errorf("Forbidden")
		}
		result = global.DBEngine.Where("ID = ?", input.Serverid).Delete(&model.Server{})
		if result.Error != nil {
			return nil, result.Error
		}
		resp.Body.Message = "success"
		return resp, nil
	})

}
