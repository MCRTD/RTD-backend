package global

import (
	storage_go "github.com/supabase-community/storage-go"
	"gorm.io/gorm"
)

var (
	DBEngine *gorm.DB
	S3Client *storage_go.Client
)
