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
	Username      string        `gorm:"not null"`
	Email         string        `gorm:"not null"`
	Password      string        `gorm:"not null" json:"Password,omitempty"`
	DiscordID     *int          `gorm:"default:null"`
	Description   string        `gorm:"default:null"`
	Avatar        string        `gorm:"default:null"`
	AvatarPath    string        `gorm:"default:null"`
	JoinedTime    time.Time     `gorm:"not null"`
	LasttimeLogin time.Time     `gorm:"not null"`
	Admin         bool          `gorm:"not null default:false"`
	SocialID      uint          `gorm:"not null"`
	Social        Social        `gorm:"foreignKey:SocialID"`
	Litematicas   []*Litematica `gorm:"many2many:litematica_creators;"`
	Groups        []*Group      `gorm:"many2many:user_groups;"`
	Servers       []*Server     `gorm:"many2many:user_servers;"`
}

type Comment struct {
	gorm.Model
	Context      string `gorm:"not null"`
	UserID       uint   `gorm:"not null"`
	User         User   `gorm:"foreignKey:UserID"`
	LitematicaID uint   `gorm:"not null"`
}

type Image struct {
	gorm.Model
	ImageName    string `gorm:"not null"`
	ImagePath    string `gorm:"not null"`
	LitematicaID uint   `gorm:"not null"`
}

type LitematicaObj struct {
	gorm.Model
	ObjFilePath string
	MtlFilePath string
	ZipFilePath string
}

type LitematicaFile struct {
	gorm.Model
	Size            int           `gorm:"not null"`
	Description     string        `gorm:"not null"`
	FileType        string        `gorm:"not null"` // litematica, schematic, world
	FileExtension   string        `gorm:"not null"` // zip, litematica, schematic, tar.gz...
	FileName        string        `gorm:"not null"`
	FilePath        string        `gorm:"not null"`
	DownloadCount   int           `gorm:"not null default:0"`
	ReleaseDate     time.Time     `gorm:"not null"`
	LitematicaID    int           `gorm:"not null"`
	LitematicaObjID uint          `gorm:"not null"`
	LitematicaObj   LitematicaObj `gorm:"foreignKey:LitematicaObjID"`
}

type Litematica struct {
	gorm.Model
	LitematicaName string  `gorm:"not null"`
	Version        string  `gorm:"not null"`
	Description    string  `gorm:"not null"`
	Tags           string  `gorm:"not null"`
	Vote           int     `gorm:"not null Default:0"`
	GroupID        *uint   `gorm:"default:null"`
	Group          *Group  `gorm:"foreignKey:GroupID"`
	ServerID       *uint   `gorm:"default:null"`
	Server         *Server `gorm:"foreignKey:ServerID"`
	Creators       []*User `gorm:"many2many:litematica_creators;"`
	Images         []Image
	Files          []LitematicaFile
	Comments       []Comment
}

type Group struct {
	gorm.Model
	GroupName   string  `gorm:"not null"`
	Description string  `gorm:"not null"`
	Avatar      string  `gorm:"not null"`
	SocialID    uint    `gorm:"not null"`
	Social      Social  `gorm:"foreignKey:SocialID"`
	Users       []*User `gorm:"many2many:user_groups;"`
}

type Server struct {
	gorm.Model
	ServerName  string  `gorm:"not null"`
	Description string  `gorm:"not null"`
	Avatar      string  `gorm:"not null"`
	SocialID    uint    `gorm:"not null"`
	Social      Social  `gorm:"foreignKey:SocialID"`
	Users       []*User `gorm:"many2many:user_servers;"`
}

type ResourcePack struct {
	gorm.Model
	Name string `gorm:"not null"`
	Path string `gorm:"not null"`
}

type LitematicaServer struct {
	gorm.Model
	ServerName string `gorm:"not null"`
	ServerIP   string `gorm:"not null"`
	Port       int    `gorm:"not null"`
	Password   string `gorm:"not null"`
}
