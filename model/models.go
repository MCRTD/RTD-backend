package model

import (
	"time"

	"gorm.io/gorm"
)

type Social struct {
	gorm.Model
	Youtube   string
	Twitter   string
	Discord   string
	Instagram string
	Telegram  string
	Github    string
}

type User struct {
	gorm.Model
	Username      string `gorm:"not null"`
	Email         string `gorm:"not null"`
	Password      string `gorm:"not null"`
	DiscordID     *int
	Description   string
	Avatar        string
	JoinedTime    time.Time     `gorm:"not null"`
	LasttimeLogin time.Time     `gorm:"not null"`
	Admin         int           `gorm:"not null"`
	SocialID      int           `gorm:"not null"`
	Social        Social        `gorm:"foreignKey:SocialID"`
	Litematicas   []*Litematica `gorm:"many2many:litematica_creators;"`
	Groups        []*Group      `gorm:"many2many:user_groups;"`
	Servers       []*Server     `gorm:"many2many:user_servers;"`
}

type Comment struct {
	gorm.Model
	Context string `gorm:"not null"`
	UserID  int    `gorm:"not null"`
	User    User   `gorm:"foreignKey:UserID"`
}

type Image struct {
	gorm.Model
	ImageName string `gorm:"type:not null"`
	ImagePath string `gorm:"type:not null"`
}

type LitematicaObj struct {
	gorm.Model
	FilePath    string `gorm:"type:not null"`
	ZipFilePath string `gorm:"type:not null"`
}

type LitematicaFile struct {
	gorm.Model
	Size            int           `gorm:"not null"`
	Description     string        `gorm:"type:not null"`
	FileName        string        `gorm:"type:not null"`
	FilePath        string        `gorm:"type:not null"`
	DownloadCount   int           `gorm:"not null"`
	ReleaseDate     time.Time     `gorm:"not null"`
	LitematicaObjID int           `gorm:"not null"`
	LitematicaObj   LitematicaObj `gorm:"foreignKey:LitematicaObjID"`
}

type Litematica struct {
	gorm.Model
	LitematicaName string `gorm:"not null"`
	Version        string `gorm:"not null"`
	Description    string `gorm:"not null"`
	Tags           string `gorm:"not null"`
	Vote           int    `gorm:"not null"`
	GroupID        int
	Group          Group `gorm:"foreignKey:GroupID"`
	ServerID       int
	Server         Server  `gorm:"foreignKey:ServerID"`
	Creators       []*User `gorm:"many2many:litematica_creators;"`
	Images         []Image
	Files          []LitematicaFile
	Comments       []Comment
}

type Group struct {
	gorm.Model
	GroupName   string  `gorm:"type:not null"`
	Description string  `gorm:"type:not null"`
	Avatar      string  `gorm:"type:not null"`
	SocialID    int     `gorm:"not null"`
	Social      Social  `gorm:"foreignKey:SocialID"`
	Users       []*User `gorm:"many2many:user_groups;"`
}

type Server struct {
	gorm.Model
	ServerName  string  `gorm:"type:not null"`
	Description string  `gorm:"type:not null"`
	Avatar      string  `gorm:"type:not null"`
	SocialID    int     `gorm:"not null"`
	Social      Social  `gorm:"foreignKey:SocialID"`
	Users       []*User `gorm:"many2many:user_servers;"`
}

type ResourcePack struct {
	gorm.Model
	Name string `gorm:"type:not null"`
	Path string `gorm:"type:not null"`
}

type LitematicaServer struct {
	gorm.Model
	ServerName string `gorm:"type:not null"`
	ServerIP   string `gorm:"type:not null"`
	Port       int    `gorm:"not null"`
	Password   string `gorm:"type:not null"`
}
