package routes

import (
	"RTD-backend/global"
	"RTD-backend/model"
	"context"
	"mime/multipart"

	"github.com/danielgtaylor/huma/v2"
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
		if input.LitematicaID == "null" {
			global.DBEngine.Model(&model.Litematica{}).Find(&Litematicas)
		} else {
			global.DBEngine.Model(&model.Litematica{}).Where("ID = ?", input.LitematicaID).Find(&Litematicas)
		}
		resp.Body.Litematicas = append(resp.Body.Litematicas, Litematicas...)

		return resp, nil

	})

	huma.Register(api, huma.Operation{
		OperationID: "postlitematica",
		Method:      "POST",
		Path:        "/litematica",
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

		litematica := &model.Litematica{
			LitematicaName: input.Name,
			Version:        input.Version,
			Description:    input.Description,
			Tags:           input.Tags,
			Files: []model.LitematicaFile{
				{
					Size:        int(file.Size),
					Description: "litematica",
					FileName:    file.Filename,
					FilePath:    url.SignedURL,
					ReleaseDate: global.DBEngine.NowFunc(),
				},
			},
		}

		if input.GroupID == -1 {
			litematica.GroupID = input.GroupID
		}

		if input.ServerID == -1 {
			litematica.ServerID = input.ServerID
		}

		global.DBEngine.Create(litematica)
		resp.Body.Message = "Success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "editlitematica",
		Method:      "PATCH",
		Path:        "/litematica",
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
}
