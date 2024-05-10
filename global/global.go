package global

import (
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

var (
	DBEngine *gorm.DB
	S3Client *minio.Client
)
