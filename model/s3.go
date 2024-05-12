package model

import (
	"os"

	storage_go "github.com/supabase-community/storage-go"
)

func NewS3Client() *storage_go.Client {
	endpoint := os.Getenv("Supabaseurl")
	storageClient := storage_go.NewClient(endpoint, os.Getenv("Supabasekey"), nil)
	return storageClient
}
