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
