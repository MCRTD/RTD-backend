package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"RTD-backend/global"
	"RTD-backend/middleware"
	"RTD-backend/model"
	"RTD-backend/routes"
	"RTD-backend/setting"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humagin"
	"github.com/gin-contrib/cors"

	_ "github.com/joho/godotenv/autoload"
	storage_go "github.com/supabase-community/storage-go"
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

	corsconfig := cors.DefaultConfig()
	corsconfig.AllowCredentials = true
	corsconfig.AllowOrigins = []string{"http://localhost:5173", "http://localhost:8888", "https://rtdweb.zeabur.app", "https://mcrtd.whitecloud.life"}
	router.Use(cors.New(
		corsconfig,
	))

	group := router.Group("/api")
	config := huma.DefaultConfig("My API", "1.0.0")
	config.Servers = []*huma.Server{{URL: "http://localhost:8888/api"}}

	api := humagin.NewWithGroup(router, group, config)
	// api.UseMiddleware(middleware.Corsfunc)
	api.UseMiddleware(middleware.ReflashHandler)
	routes.Helloworld(api)
	routes.User(api)
	routes.Node(api)
	routes.Litematica(api)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	log.Println("Server started on http://127.0.0.1:" + port)
	log.Println("Docs  http://127.0.0.1:" + port + "/api/docs")
	http.ListenAndServe(":"+port, router)
}

func Syncddb() {
	err := global.DBEngine.AutoMigrate(
		&model.Social{},
		&model.User{},
		&model.Comment{},
		&model.Image{},
		&model.LitematicaObj{},
		&model.LitematicaFile{},
		&model.LitematicaVote{},
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
	global.S3Client = model.NewS3Client()
	if err != nil {
		return err
	}
	// Syncddb()
	_, error := global.S3Client.GetBucket("litematica")
	if error != nil {
		_, err := global.S3Client.CreateBucket("litematica", storage_go.BucketOptions{
			Public: true,
		})
		if err != nil {
			return err
		}
	}

	_, error = global.S3Client.GetBucket("images")
	if error != nil {
		_, err := global.S3Client.CreateBucket("images", storage_go.BucketOptions{
			Public: true,
		})
		if err != nil {
			return err
		}
	}
	_, error = global.S3Client.GetBucket("texturepack")
	if error != nil {
		_, err := global.S3Client.CreateBucket("texturepack", storage_go.BucketOptions{
			Public: true,
		})
		if err != nil {
			return err
		}
	}
	_, error = global.S3Client.GetBucket("obj")
	if error != nil {
		_, err := global.S3Client.CreateBucket("obj", storage_go.BucketOptions{
			Public: true,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
