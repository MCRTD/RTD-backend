package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"RTD-backend/global"
	"RTD-backend/model"
	"RTD-backend/routes"
	"RTD-backend/setting"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"

	_ "github.com/joho/godotenv/autoload"
)

type GreetingOutput struct {
	Body struct {
		Message string `json:"message" example:"Hello, world!" doc:"Greeting message"`
	}
}

func init() {
	err := SetupDB()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Database connected")
	}
}

func main() {
	router := gin.Default()

	group := router.Group("/api")
	config := huma.DefaultConfig("My API", "1.0.0")
	config.Servers = []*huma.Server{{URL: "http://localhost:8888/api"}}
	api := humagin.NewWithGroup(router, group, config)
	routes.Helloworld(api)
	routes.User(api)
	routes.Node(api)
	routes.Litematica(api)
	log.Println("Server started on http://127.0.0.1:8888")
	log.Println("Docs  http://127.0.0.1:8888/api/docs")
	http.ListenAndServe("127.0.0.1:8888", router)
}

func Syncddb() {
	err := global.DBEngine.AutoMigrate(
		&model.Social{},
		&model.User{},
		&model.Comment{},
		&model.Image{},
		&model.LitematicaObj{},
		&model.LitematicaFile{},
		&model.Litematica{},
		&model.Group{},
		&model.Server{},
		&model.ResourcePack{},
		&model.LitematicaServer{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

func SetupDB() error {
	var err error

	DBSetting := setting.GetDatabaseSetting()
	global.DBEngine, err = model.NewDBEngine(DBSetting)
	if err != nil {
		return err
	}
	Syncddb()

	return nil
}
