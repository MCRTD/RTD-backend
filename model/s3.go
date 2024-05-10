package model

import (
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewS3Client() *minio.Client {
	endpoint := "https://swomajagzechqsnrpefx.supabase.co/storage/v1/s3"
	accessKeyID := ""
	secretAccessKey := ""
	useSSL := true
	// bucketName := "litematica"
	// location := "ap-northeast-1"

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}
	return minioClient

}
