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
		OperationID: "getnode",
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
}
