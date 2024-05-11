package api

import (
	"RTD-backend/global"
	"RTD-backend/model"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

type Nodetexturepack struct {
	Texturepacks []string
}

type NodeOBJMessage struct {
	Message string
}

func MakeOBJ(id string, texturepack string, name string) NodeOBJMessage {
	var servers []model.LitematicaServer
	global.DBEngine.Model(&model.LitematicaServer{}).Find(&servers)
	server := servers[0]
	req, err := http.NewRequest("POST", server.ServerIP+":"+string(rune(server.Port))+"/litematica/upload", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("texturepack", texturepack)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()
	global.S3Client.UpdateFile("litematica", name+".obj", resp.Body)

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
