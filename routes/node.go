package routes

import (
	"RTD-backend/global"
	"RTD-backend/model"
	"context"
	"fmt"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
)

type NodeInfo struct {
	Name string
	IP   string
	Port string
	ID   uint
}

type NodeOutput struct {
	Body struct {
		Servers []NodeInfo `json:"servers"`
	}
}

type PatchNode struct {
	ID       uint   `json:"ID"`
	Name     string `json:"Name"`
	IP       string `json:"Ip"`
	Port     int    `json:"Port"`
	Password string `json:"Password"`
}

func Node(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getnode",
		Method:      "GET",
		Path:        "/node",
	}, func(ctx context.Context, input *struct {
		Nodename string `header:"Nodename" example:"server1" doc:"the name of server"`
	}) (*NodeOutput, error) {
		resp := &NodeOutput{}
		var servers []model.LitematicaServer
		if input.Nodename == "" {
			global.DBEngine.Model(&model.LitematicaServer{}).Find(&servers)
		} else {
			global.DBEngine.Model(&model.LitematicaServer{}).Where("ID = ?", input.Nodename).Find(&servers)
		}

		for _, server := range servers {
			resp.Body.Servers = append(resp.Body.Servers, NodeInfo{
				Name: server.ServerName,
				IP:   server.ServerIP,
				Port: strconv.Itoa(server.Port),
				ID:   server.ID,
			})
		}
		return resp, nil

	})

	huma.Register(api, huma.Operation{
		OperationID: "addnode",
		Method:      "POST",
		Path:        "/node",
	}, func(ctx context.Context, input *struct {
		Name     string `header:"name" example:"server1" doc:"Server name"`
		Ip       string `header:"ip" example:"127.0.0.1" doc:"IP"`
		Port     int    `header:"port" example:"8888" doc:"Port"`
		Password string `header:"password" example:"password" doc:"Password."`
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		global.DBEngine.Create(&model.LitematicaServer{
			ServerName: input.Name,
			ServerIP:   input.Ip,
			Port:       input.Port,
			Password:   input.Password,
		})
		resp.Body.Message = "Add node success!"
		return resp, nil

	})

	huma.Register(api, huma.Operation{
		OperationID: "patchnode",
		Method:      "PATCH",
		Path:        "/node",
	}, func(ctx context.Context, input *struct {
		Body PatchNode
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		fmt.Println(input.Body)
		global.DBEngine.Model(&model.LitematicaServer{}).Where("ID = ?", input.Body.ID).Updates(&map[string]interface{}{
			"ServerName": input.Body.Name,
			"ServerIP":   input.Body.IP,
			"Port":       input.Body.Port,
			"Password":   input.Body.Password,
		})
		resp.Body.Message = "Add node success!"
		return resp, nil

	})
}
