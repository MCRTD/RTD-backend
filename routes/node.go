package routes

import (
	"RTD-backend/global"
	"RTD-backend/lapi"
	"RTD-backend/model"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
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

type TexturepackResponse struct {
	Body struct {
		Texturepacks []string `json:"texturepacks"`
	}
}

func Node(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getnode",
		Method:      "GET",
		Path:        "/node",
	}, func(ctx context.Context, input *struct {
		Nodeid string `header:"Nodeid" example:"1" doc:"the id of server"`
	}) (*NodeOutput, error) {
		resp := &NodeOutput{}
		var servers []model.LitematicaServer
		if input.Nodeid == "" {
			global.DBEngine.Model(&model.LitematicaServer{}).Find(&servers)
		} else {
			global.DBEngine.Model(&model.LitematicaServer{}).Where("ID = ?", input.Nodeid).Find(&servers)
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

	huma.Register(api, huma.Operation{
		OperationID: "deletenode",
		Method:      "DELETE",
		Path:        "/node",
	}, func(ctx context.Context, input *struct {
		NodeID uint `header:"NodeID" example:"1" doc:"NodeID"`
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		node := model.LitematicaServer{Model: gorm.Model{ID: input.NodeID}}
		global.DBEngine.Delete(&node)
		resp.Body.Message = "Delete node success!"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "gettexturepack",
		Method:      "GET",
		Path:        "/node/texturepack",
	}, func(ctx context.Context, input *struct {
		Nodeid string `header:"Nodeid" example:"1" doc:"the id of server" required:"true"`
	}) (*TexturepackResponse, error) {

		texturepacks := lapi.GetResourcePacksFromNode(input.Nodeid)
		fmt.Println(texturepacks)
		resp := &TexturepackResponse{}
		resp.Body.Texturepacks = texturepacks
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "addnodetexturepack",
		Method:      "POST",
		Path:        "/node/texturepack",
	}, func(ctx context.Context, input *struct {
		Name    string `header:"Name" example:"vanilla" doc:"texturepack name"`
		RawBody multipart.Form
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		file := input.RawBody.File["texturepack"][0]
		filedata, err := file.Open()
		if err != nil {
			resp.Body.Message = "Failed in opening file"
			log.Println(err)
			return resp, err
		}
		if strings.Split(file.Filename, ".")[len(strings.Split(file.Filename, "."))-1] != "zip" {
			resp.Body.Message = "must be a zip file"
			log.Println(err)
			return resp, err
		}

		go func() {
			err = lapi.UploadTexturePackToNode(input.Name, filedata)
			if err != nil {
				log.Println(err)
			}
		}()
		resp.Body.Message = "File is being uploaded"
		return resp, nil
	})

}
