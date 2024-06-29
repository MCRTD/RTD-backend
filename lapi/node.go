package lapi

import (
	"RTD-backend/global"
	"RTD-backend/model"
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"

	storage_go "github.com/supabase-community/storage-go"
)

type Nodetexturepack struct {
	Texturepacks []string
}

type NodeOBJMessage struct {
	Message string
}

func UploadTexturePackToNode(texturepack string, file io.Reader) error {
	var servers []model.LitematicaServer
	global.DBEngine.Model(&model.LitematicaServer{}).Find(&servers)
	if len(servers) == 0 {
		log.Print("No server available")
		return nil
	}
	diruuid := uuid.New()
	dir, err := os.MkdirTemp("", "texturepack"+diruuid.String())
	if err != nil {
		log.Println(err)
		return err
	}
	defer os.RemoveAll(dir)

	buff := bytes.NewBuffer([]byte{})
	size, err := io.Copy(buff, file)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(buff.Bytes())
	zipReader, err := zip.NewReader(reader, size)

	if err != nil {
		log.Println(err)
		return err
	}

	for _, f := range zipReader.File {
		filePath := filepath.Join(dir, f.Name)
		fmt.Println("unzipping file ", filePath)

		if !strings.HasPrefix(filePath, filepath.Clean(dir)+string(os.PathSeparator)) {
			fmt.Println("invalid file path")
			return fmt.Errorf("%s: illegal file path", filePath)
		}
		if f.FileInfo().IsDir() {
			fmt.Println("creating directory...")
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			panic(err)
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}

		fileInArchive, err := f.Open()
		if err != nil {
			panic(err)
		}

		if _, err := io.Copy(dstFile, fileInArchive); err != nil {
			panic(err)
		}

		dstFile.Close()
		fileInArchive.Close()
	}

	texturesDir := filepath.Join(dir, "assets", "minecraft", "textures")
	var wg sync.WaitGroup
	sem := make(chan struct{}, 20)

	err = filepath.Walk(texturesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("遇到錯誤: %v", err)
			return err
		}

		if info.IsDir() {
			log.Printf("跳過目錄: %s", path)
			return nil
		}

		log.Printf("處理文件: %s", path)

		fileData, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if strings.HasSuffix(relPath, ".json") || strings.HasSuffix(relPath, ".fsh") || strings.HasSuffix(relPath, ".vsh") {
			return nil
		}

		log.Printf("上傳文件: %s", relPath)

		wg.Add(1)
		sem <- struct{}{}
		go func(path, relPath string, fileData []byte) {
			defer wg.Done()
			defer func() { <-sem }()

			relPath = strings.TrimPrefix(relPath, "assets\\minecraft\\textures\\")
			relPath = strings.Replace(filepath.Join(texturepack, relPath), "\\", "/", -1)
			contentType := http.DetectContentType(fileData)
			filesetting := storage_go.FileOptions{
				ContentType: &contentType,
			}
			_, err = global.S3Client.UploadFile("texturepack", relPath, bytes.NewReader(fileData), filesetting)

			if err != nil {
				log.Printf("上傳文件時遇到錯誤: %v", err)
				return
			}
		}(path, relPath, fileData)

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	for _, server := range servers {
		req, err := http.NewRequest("POST", server.ServerIP+":"+strconv.Itoa(server.Port)+"/texturepack/upload?texturepackname="+texturepack, nil)
		if err != nil {
			log.Println(err)
			return err
		}
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		part, err := writer.CreateFormFile("file", texturepack+".zip")
		if err != nil {
			log.Print(err)
			return err
		}

		fileContent, err := io.ReadAll(file)
		if err != nil {
			log.Print(err)
			return err
		}

		_, err = part.Write(fileContent)
		if err != nil {
			log.Print(err)
			return err
		}

		err = writer.Close()
		if err != nil {
			log.Print(err)
			return err
		}

		req.Body = io.NopCloser(body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return err
		}
		if resp != nil {
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				log.Println("Received non-OK response code")
				return fmt.Errorf("received non-OK response code")
			}
		} else {
			return fmt.Errorf("no response received")
		}

	}

	return nil

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

	diruuid := uuid.New()
	dir, err := os.MkdirTemp("", "litematicaobj"+diruuid.String())
	if err != nil {
		log.Println(err)
		return NodeOBJMessage{Message: "Error creating temp directory"}
	}
	defer os.RemoveAll(dir)

	buff := bytes.NewBuffer([]byte{})
	size, err := io.Copy(buff, resp.Body)
	if err != nil {
		log.Println(err)
		return NodeOBJMessage{Message: "Error copying response"}
	}
	reader := bytes.NewReader(buff.Bytes())
	zipReader, err := zip.NewReader(reader, size)

	// 將.mtl與.obj檔案上傳至S3

	for _, f := range zipReader.File {
		filePath := filepath.Join(dir, f.Name)
		fmt.Println("unzipping file ", filePath)

		if strings.HasSuffix(f.Name, ".mtl") || strings.HasSuffix(f.Name, ".obj") {
			fileInArchive, err := f.Open()
			if err != nil {
				log.Println("Error opening file:", err)
				return NodeOBJMessage{Message: "Error opening file"}
			}
			defer fileInArchive.Close()
			_, fileerr := global.S3Client.UploadFile("obj", name+".obj", fileInArchive, storage_go.FileOptions{})
			if fileerr != nil {
				log.Println("Error uploading file:", fileerr)
				return NodeOBJMessage{Message: "Error uploading file"}
			}
		}
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

func GetResourcePacksFromNode(id string) []string {
	var servers []model.LitematicaServer
	global.DBEngine.Model(&model.LitematicaServer{}).Where("id = ?", id).Find(&servers)
	var results []string
	for _, url := range servers {
		req, err := http.NewRequest("GET", url.ServerIP+":"+strconv.Itoa(url.Port)+"/texturepack/list", nil)
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
		results = append(results, result.Texturepacks...)
	}
	return results

}
