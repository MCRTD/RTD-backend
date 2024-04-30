package routes

import (
	"context"

	"github.com/danielgtaylor/huma/v2"
)

type GreetingOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
	}
}

func Helloworld(api huma.API) {
	// just return hello world
	// /helloworld/

	huma.Get(api, "/helloworld", func(ctx context.Context, input *struct{}) (*GreetingOutput, error) {
		resp := &GreetingOutput{}
		resp.Body.Message = "Hello, world!"
		return resp, nil
	})
}
