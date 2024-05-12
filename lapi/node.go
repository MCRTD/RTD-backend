package lapi

import (
	"RTD-backend/global"
	"RTD-backend/model"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	storage_go "github.com/supabase-community/storage-go"
)

type Nodetexturepack struct {
	Texturepacks []string
}

type NodeOBJMessage struct {
	Message string
}

func MakeOBJ(url string, texturepack string, name string, id int) NodeOBJMessage {
	var servers []model.LitematicaServer
	global.DBEngine.Model(&model.LitematicaServer{}).Find(&servers)
	if len(servers) == 0 {
		log.Print("No server available")
		return NodeOBJMessage{Message: "No server available"}
	}
	server := servers[0]

	req, err := http.NewRequest("POST", server.ServerIP+":"+strconv.Itoa(server.Port)+"/litematica/upload", nil)
	if err != nil {
		log.Println(err)
		return NodeOBJMessage{Message: "Error creating request"}
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("file", url)
	_ = writer.WriteField("texturepack", texturepack)
	err = writer.Close()
	if err != nil {
		log.Print(err)
		return NodeOBJMessage{Message: "Error closing writer"}
	}

	req.Body = io.NopCloser(body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return NodeOBJMessage{Message: "Error making request"}
	}
	if resp != nil {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Println("Received non-OK response code")
			return NodeOBJMessage{Message: "Received non-OK response code"}
		}
	} else {
		return NodeOBJMessage{Message: "No response received"}
	}

	contentType := "application/zip"
	_, fileerr := global.S3Client.UploadFile("litematica", name+".zip", resp.Body, storage_go.FileOptions{
		ContentType: &contentType,
	})
	if fileerr != nil {
		log.Println(fileerr)
		return NodeOBJMessage{Message: "Error uploading file"}
	}

	resp.Body.Close()

	global.DBEngine.Model(&model.LitematicaFile{}).Where("ID = ?", id).Update("LitematicaObj", &model.LitematicaObj{ZipFilePath: global.S3Client.GetPublicUrl("litematica", name+".zip").SignedURL})

	return NodeOBJMessage{Message: "Success"}
}

func GetResourcePacksFromNode(id string) []Nodetexturepack {
	var servers []model.LitematicaServer
	global.DBEngine.Model(&model.LitematicaServer{}).Where("id = ?", id).Find(&servers)
	var results []Nodetexturepack
	for _, url := range servers {
		req, err := http.NewRequest("GET", url.ServerIP+":"+string(rune(url.Port))+"/texturepack/list", nil)
		if err != nil {
			log.Fatal(err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		var result Nodetexturepack
		err = json.Unmarshal(body, &result)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, result)
	}
	return results

}
