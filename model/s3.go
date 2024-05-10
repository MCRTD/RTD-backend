package model

import (
	"os"

	storage_go "github.com/supabase-community/storage-go"
)

func NewS3Client() *storage_go.Client {
	endpoint := "swomajagzechqsnrpefx.supabase.co/storage/v1"
	storageClient := storage_go.NewClient(endpoint, os.Getenv("Supabasekey"), nil)
	return storageClient
}
