package routes

import (
	"RTD-backend/global"
	"RTD-backend/lapi"
	"RTD-backend/middleware"
	"RTD-backend/model"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	storage_go "github.com/supabase-community/storage-go"

	"github.com/danielgtaylor/huma/v2"
	"gorm.io/gorm"
)

type LitematicaOutput struct {
	Body struct {
		Litematicas []model.Litematica `json:"litematicas"`
	}
}

type PostObj struct {
	FileID     int    `json:"FileID"`
	Texurepack string `json:"Texurepack"`
}

type LitematicaID struct {
	LitematicaID int `json:"LitematicaID"`
}

type VoteInput struct {
	LitematicaID int `json:"LitematicaID"`
}

func Litematica(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "getlitematica",
		Method:      "GET",
		Path:        "/litematica",
	}, func(ctx context.Context, input *struct {
		LitematicaID string `query:"LitematicaID" example:"1" doc:"LitematicaID"`
	}) (*LitematicaOutput, error) {
		resp := &LitematicaOutput{}
		var Litematicas []model.Litematica

		query := global.DBEngine.
			Preload("Files.LitematicaObj").
			Preload("Creators.Social", func(db *gorm.DB) *gorm.DB {
				return db.Where("id IS NOT NULL")
			}).
			Preload("Creators", func(db *gorm.DB) *gorm.DB {
				return db.Where("id IS NOT NULL")
			}).
			Model(&model.Litematica{})

		if input.LitematicaID == "" {
			query.Find(&Litematicas)
		} else {
			query.Where("ID = ?", input.LitematicaID).Find(&Litematicas)
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
		FileType    string `header:"FileType" example:"litematica" doc:"litematica, schematic, world"`
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

		if objfile.ID == 0 {
			resp.Body.Message = "Failed in creating objfile"
			return resp, huma.NewError(400, "Failed in creating objfile")
		}

		if input.Name == "" {
			resp.Body.Message = "Name is empty"
			return resp, huma.NewError(400, "Name is empty")
		}

		fileext := file.Filename[strings.LastIndex(file.Filename, ".")+1:]
		if strings.Contains(file.Filename, ".tar.gz") {
			fileext = "tar.gz"
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
					FileType:        input.FileType,
					FileExtension:   fileext,
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
			groupID := uint(input.GroupID)
			litematica.GroupID = &groupID
		}

		if input.ServerID != -1 {
			serverID := uint(input.ServerID)
			litematica.ServerID = &serverID
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
		Body LitematicaID
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}

		litematica := model.Litematica{Model: gorm.Model{ID: uint(input.Body.LitematicaID)}}

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
		OperationID: "addlitematicaimage",
		Method:      "POST",
		Path:        "/litematica/image",
	}, func(ctx context.Context, input *struct {
		RawBody multipart.Form
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		litematicaID, err := strconv.ParseUint(input.RawBody.Value["LitematicaID"][0], 10, 64)
		if err != nil {
			resp.Body.Message = "Invalid LitematicaID"
			return resp, huma.Error400BadRequest("Invalid LitematicaID")
		}
		files := input.RawBody.File["image"]
		for _, file := range files {
			if file.Size > 1024*1024*35 {
				resp.Body.Message = "File is too large"
				return resp, huma.NewError(413, "File is too large")
			}
		}
		for _, file := range files {
			filedata, err := file.Open()
			if err != nil {
				resp.Body.Message = "Failed"
				return resp, huma.Error400BadRequest("Failed")
			}
			parts := strings.Split(file.Filename, ".")
			newfilename := uuid.New().String() + "." + parts[len(parts)-1]
			buffer := make([]byte, 512)
			_, err = filedata.Read(buffer)
			if err != nil {
				fmt.Println("Failed to read file:", err)
				resp.Body.Message = "Failed"
				return resp, err
			}
			_, err = filedata.Seek(0, io.SeekStart)
			if err != nil {
				fmt.Println("Failed to seek file:", err)
				resp.Body.Message = "Failed"
				return resp, err
			}
			filetype := http.DetectContentType(buffer)
			_, err = global.S3Client.UploadFile("images", newfilename, filedata, storage_go.FileOptions{
				ContentType: &filetype,
			})
			if err != nil {
				resp.Body.Message = "Failed"
				return resp, huma.Error400BadRequest("Failed")
			}
			url := global.S3Client.GetPublicUrl("images", newfilename)
			litematicaImage := &model.Image{
				LitematicaID: uint(litematicaID),
				ImageName:    newfilename,
				ImagePath:    url.SignedURL,
			}
			success := global.DBEngine.Create(litematicaImage)
			if success.Error != nil {
				resp.Body.Message = "Failed"
				return resp, huma.Error400BadRequest("Failed")
			}

		}
		resp.Body.Message = "Success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "deletelitematicaimage",
		Method:      "DELETE",
		Path:        "/litematica/image",
	}, func(ctx context.Context, input *struct {
		ImageID uint `header:"ImageID" example:"1" doc:"ImageID"`
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}

		global.DBEngine.Delete(&model.Image{Model: gorm.Model{ID: input.ImageID}})

		resp.Body.Message = "Success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "addlitematicafile",
		Method:      "POST",
		Path:        "/litematica/file",
	}, func(ctx context.Context, input *struct {
		Body    LitematicaID
		RawBody multipart.Form
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
			LitematicaID: input.Body.LitematicaID,
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
		Body PostObj
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		file := model.LitematicaFile{}
		global.DBEngine.Model(&model.LitematicaFile{}).Where("ID = ?", input.Body.FileID).Find(&file)

		if file.ID == 0 {
			resp.Body.Message = "File not found"
			return nil, huma.NewError(400, "File not found")
		}

		go func() {
			lapi.MakeOBJ(file.FilePath, input.Body.Texurepack, file.FileName, input.Body.FileID)
		}()

		resp.Body.Message = "Success"
		return resp, nil
	})

	huma.Register(api, huma.Operation{
		OperationID: "voteLitematica",
		Method:      "POST",
		Path:        "/litematica/vote",
		Middlewares: huma.Middlewares{middleware.ParseToken(api)},
	}, func(ctx context.Context, input *struct {
		Body VoteInput
	}) (*NormalOutput, error) {
		resp := &NormalOutput{}
		userIDStr := ctx.Value("userid").(string)
		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			return nil, huma.NewError(400, "Invalid user ID")
		}

		tx := global.DBEngine.Begin()
		var litematica model.Litematica
		if err := tx.First(&litematica, input.Body.LitematicaID).Error; err != nil {
			tx.Rollback()
			return nil, huma.NewError(404, "Litematica not found")
		}
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			tx.Rollback()
			return nil, huma.NewError(404, "User not found")
		}
		var count int64
		if err := tx.Table("litematica_votes").
			Where("litematica_id = ? AND user_id = ?", litematica.ID, userID).
			Count(&count).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		hasVoted := count > 0

		if !hasVoted {
			if err := tx.Model(&litematica).Association("VoteUsers").Append(&user); err != nil {
				tx.Rollback()
				return nil, err
			}
			if err := tx.Model(&litematica).Update("vote", litematica.Vote+1).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			resp.Body.Message = "Vote added"
		} else {
			if err := tx.Model(&litematica).Association("VoteUsers").Delete(&user); err != nil {
				tx.Rollback()
				return nil, err
			}
			if err := tx.Model(&litematica).Update("vote", litematica.Vote-1).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
			resp.Body.Message = "Vote removed"
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		return resp, nil
	})
}
