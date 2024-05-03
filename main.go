package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"RTD-backend/routes"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
)

type GreetingOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
	}
}

func main() {
	router := gin.Default()

	group := router.Group("/api")
	config := huma.DefaultConfig("My API", "1.0.0")
	config.Servers = []*huma.Server{{URL: "http://localhost:8888/api"}}
	api := humagin.NewWithGroup(router, group, config)
	huma.Get(api, "/greeting/{name}", func(ctx context.Context, input *struct {
		Name string `path:"name" maxLength:"30" example:"world" doc:"Name to greet"`
	}) (*GreetingOutput, error) {
		resp := &GreetingOutput{}
		resp.Body.Message = fmt.Sprintf("Hello, %s!", input.Name)
		return resp, nil
	})
	routes.Helloworld(api)
	routes.Register(api)
	http.ListenAndServe("127.0.0.1:8888", router)
}
