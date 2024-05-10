package routes

import (
	"RTD-backend/global"
	"RTD-backend/model"
	"context"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
)

type NodeInfo struct {
	Name string
	IP   string
	Port string
}

type NodeOutput struct {
	Body struct {
		Servers []NodeInfo `json:"servers"`
	}
}

func Node(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getnode",
		Method:      "GET",
		Path:        "/node",
	}, func(ctx context.Context, input *struct {
		node string `header:"node" example:"user" doc:"Username." nullable:"true"`
	}) (*NodeOutput, error) {
		resp := &NodeOutput{}
		var servers []model.LitematicaServer

		if input.node == "null" {
			global.DBEngine.Model(&model.LitematicaServer{}).Find(&servers)
		} else {
			global.DBEngine.Model(&model.LitematicaServer{}).Where("ID = ?", input.node).Find(&servers)
		}

		for _, server := range servers {
			resp.Body.Servers = append(resp.Body.Servers, NodeInfo{
				Name: server.ServerName,
				IP:   server.ServerIP,
				Port: strconv.Itoa(server.Port),
			})
		}
		return resp, nil

	})

	huma.Register(api, huma.Operation{
		OperationID: "addnode",
		Method:      "POST",
		Path:        "/node",
	}, func(ctx context.Context, input *struct {
		Name     string `header:"node" example:"server1" doc:"Server name"`
		ip       string `header:"ip" example:"127.0.0.1" doc:"IP"`
		Port     int    `header:"port" example:"8888" doc:"Port"`
		Password string `header:"password" example:"password" doc:"Password."`
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		global.DBEngine.Create(&model.LitematicaServer{
			ServerName: input.Name,
			ServerIP:   input.ip,
			Port:       input.Port,
			Password:   input.Password,
		})
		resp.Body.Message = "Add node success!"
		return resp, nil

	})
}
