package routes

import (
	"RTD-backend/global"
	"RTD-backend/model"
	"context"

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
		LitematicaID string `header:"LitematicaID" example:"1" doc:"LitematicaID" nullable:"true"`
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
		GroupID     string `header:"GroupID" example:"1" doc:"GroupID" nullable:"true"`
		ServerID    string `header:"ServerID" example:"1" doc:"ServerID" nullable:"true"`
		File        []byte
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		resp.Body.Message = "Success"
		return resp, nil

	})
}
