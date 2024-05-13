package routes

import (
	"RTD-backend/global"
	"RTD-backend/lapi"
	"RTD-backend/middleware"
	"RTD-backend/model"
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

type LitematicaOutput struct {
	Body struct {
		Litematicas []model.Litematica `json:"servers"`
	}
}

func Litematica(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getlitematica",
		Method:      "GET",
		Path:        "/litematica",
	}, func(ctx context.Context, input *struct {
		LitematicaID string `header:"LitematicaID" example:"1" doc:"LitematicaID"`
	}) (*LitematicaOutput, error) {
		resp := &LitematicaOutput{}
		var Litematicas []model.Litematica
		if input.LitematicaID == "" {
			global.DBEngine.Preload("Files.LitematicaObj").Preload("Creators").Model(&model.Litematica{}).Find(&Litematicas)
		} else {
			global.DBEngine.Preload("Files.LitematicaObj").Preload("Creators").Model(&model.Litematica{}).Where("ID = ?", input.LitematicaID).Find(&Litematicas)
		}
		for i := range Litematicas {
			for j := range Litematicas[i].Creators {
				Litematicas[i].Creators[j].Password = ""
			}
		}
		resp.Body.Litematicas = append(resp.Body.Litematicas, Litematicas...)

		return resp, nil

	})

	huma.Register(api, huma.Operation{
		OperationID: "postlitematica",
		Method:      "POST",
		Path:        "/litematica",
		Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		Name        string `header:"Name" example:"litematica" doc:"Name"`
		Version     string `header:"Version" example:"1.0" doc:"Version"`
		Description string `header:"Description" example:"litematica" doc:"Description"`
		Tags        string `header:"Tags" example:"litematica" doc:"Tags"`
		GroupID     int    `header:"GroupID" example:"1" doc:"GroupID"`
		ServerID    int    `header:"ServerID" example:"1" doc:"ServerID"`
		RawBody     multipart.Form
	}) (*NormalOutput, error) {

		resp := &NormalOutput{}
		fmt.Println(input)

		file := input.RawBody.File["litematica"][0]

		if file == nil {
			resp.Body.Message = "File is empty"
			return resp, huma.NewError(400, "File is empty")
		}
		filedata, err := file.Open()
		if err != nil {
			resp.Body.Message = "Failed in opening file"
			log.Println(err)
			return resp, err
		}

		_, err = global.S3Client.UploadFile("litematica", file.Filename, filedata)
		if err != nil {
			if err.Error() == "The resource already exists" {
				resp.Body.Message = "File already exists"
				return resp, huma.NewError(400, "File already exists")
			}
			resp.Body.Message = "Failed in uploading file"
			log.Println(err)
			return resp, err
		}
		url := global.S3Client.GetPublicUrl("litematica", file.Filename)

		objfile := &model.LitematicaObj{}
		global.DBEngine.Create(objfile)

		if input.Name == "" {
			resp.Body.Message = "Name is empty"
			return resp, huma.NewError(400, "Name is empty")
		}

		litematica := &model.Litematica{
			LitematicaName: input.Name,
			Version:        input.Version,
			Description:    input.Description,
			Tags:           input.Tags,
			Creators:       []*model.User{},
			Files: []model.LitematicaFile{
				{
					Size:            int(file.Size),
					Description:     input.Description,
					FileName:        file.Filename,
					FilePath:        url.SignedURL,
					ReleaseDate:     global.DBEngine.NowFunc(),
					LitematicaObjID: objfile.ID,
				},
			},
		}
		me := &model.User{}
		global.DBEngine.Model(&model.User{}).Where("ID = ?", ctx.Value("userid")).First(me)
		litematica.Creators = append(litematica.Creators, me)

		if input.GroupID != -1 {
			litematica.GroupID = uint(input.GroupID)
		}

		if input.ServerID != -1 {
			litematica.ServerID = uint(input.ServerID)
		}

		global.DBEngine.Create(litematica)

		global.DBEngine.Model(litematica).Association("Creators").Replace(litematica.Creators)
		resp.Body.Message = "Success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "editlitematica",
		Method:      "PATCH",
		Path:        "/litematica",
		Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		LitematicaID   int    `header:"LitematicaID" example:"1" doc:"LitematicaID"`
		LitematicaName string `header:"LitematicaName" example:"litematica" doc:"LitematicaName"`
		Version        string `header:"Version" example:"1.0" doc:"Version"`
		Description    string `header:"Description" example:"litematica" doc:"Description"`
		Tags           string `header:"Tags" example:"litematica" doc:"Tags"`
		GroupID        int    `header:"GroupID" example:"1" doc:"GroupID"`
		ServerID       int    `header:"ServerID" example:"1" doc:"ServerID"`
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}

		global.DBEngine.Model(&model.Litematica{}).Where("ID = ?", input.LitematicaID).Updates(
			map[string]interface{}{
				"LitematicaName": input.LitematicaName,
				"Version":        input.Version,
				"Description":    input.Description,
				"Tags":           input.Tags,
				"GroupID":        input.GroupID,
				"ServerID":       input.ServerID,
			})
		resp.Body.Message = "Success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "deletelitematica",
		Method:      "DELETE",
		Path:        "/litematica",
	}, func(ctx context.Context, input *struct {
		LitematicaID int `header:"LitematicaID" example:"1" doc:"LitematicaID"`
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}

		litematica := model.Litematica{Model: gorm.Model{ID: uint(input.LitematicaID)}}

		global.DBEngine.Preload("Files").Find(&litematica)

		for _, file := range litematica.Files {
			global.DBEngine.Delete(&file)
		}
		// for _, Image := range litematica.Images {
		// 	global.DBEngine.Delete(&Image)
		// }

		global.DBEngine.Model(&litematica).Association("Creators").Clear()
		global.DBEngine.Delete(&litematica)

		resp.Body.Message = "Success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "addlitematicafile",
		Method:      "POST",
		Path:        "/litematica/file",
	}, func(ctx context.Context, input *struct {
		LitematicaID int `header:"LitematicaID" example:"1" doc:"LitematicaID"`
		RawBody      multipart.Form
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}

		file := input.RawBody.File["litematica"][0]
		filedata, err := file.Open()
		if err != nil {
			resp.Body.Message = "Failed"
			return resp, nil
		}
		_, err = global.S3Client.UploadFile("litematica", file.Filename, filedata)
		if err != nil {
			resp.Body.Message = "Failed"
			return resp, nil
		}
		url := global.S3Client.GetPublicUrl("litematica", file.Filename)

		litematicaFile := &model.LitematicaFile{
			LitematicaID: input.LitematicaID,
			FileName:     file.Filename,
			FilePath:     url.SignedURL,
		}
		global.DBEngine.Create(litematicaFile)

		resp.Body.Message = "Success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "makelitematicaobj",
		Method:      "POST",
		Path:        "/litematica/obj",
	}, func(ctx context.Context, input *struct {
		FileID     int    `header:"FileID" example:"1" doc:"FileID"`
		Texurepack string `header:"Texurepack" example:"1" doc:"Texurepack"`
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		file := model.LitematicaFile{}
		global.DBEngine.Model(&model.LitematicaFile{}).Where("ID = ?", input.FileID).Find(&file)

		go func() {
			lapi.MakeOBJ(file.FilePath, input.Texurepack, file.FileName, input.FileID)
		}()

		resp.Body.Message = "Success"
		return resp, nil
	})
}
