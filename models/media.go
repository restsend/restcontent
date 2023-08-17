package models

import (
	"path/filepath"

	"github.com/restsend/carrot"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Media struct {
	BaseContent
	Size       int64  `json:"size"`
	Directory  bool   `json:"directory" gorm:"index"`
	Path       string `json:"path" gorm:"size:200;uniqueIndex:,composite:_path_name"`
	Name       string `json:"name" gorm:"size:200;uniqueIndex:,composite:_path_name"`
	Ext        string `json:"ext" gorm:"size:100"`
	Dimensions string `json:"dimensions" gorm:"size:200"` // x*y
	StorePath  string `json:"-" gorm:"size:300"`
	External   bool   `json:"external"`
	PublicUrl  string `json:"publicUrl,omitempty" gorm:"-"`
}
type MediaFolder struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	FilesCount   int64  `json:"filesCount"`
	FoldersCount int64  `json:"foldersCount"`
}

func (m *Media) BuildPublicUrls(mediaHost string, mediaPrefix string) {
	if m.Directory {
		m.PublicUrl = ""
		return
	}

	publicUrl := filepath.Join(mediaPrefix, m.Path, m.Name)
	if mediaHost != "" {
		if mediaHost[len(mediaHost)-1] == '/' {
			mediaHost = mediaHost[:len(mediaHost)-1]
		}
		publicUrl = mediaHost + publicUrl
	}
	m.PublicUrl = publicUrl

	if m.ContentType == ContentTypeImage && m.Thumbnail == "" {
		m.Thumbnail = m.PublicUrl
	}
}

func CreateFolder(db *gorm.DB, parent, name string, user *carrot.User) (string, error) {
	if parent == "" {
		parent = "/"
	}
	obj := Media{
		Path:      parent,
		Name:      name,
		Directory: true,
	}

	if user != nil {
		obj.Creator = *user
		obj.CreatorID = user.ID
	}

	fullPath := filepath.Join(parent, name)
	return fullPath, db.Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&obj).Error
}

func ListFolders(db *gorm.DB, path string) ([]MediaFolder, error) {
	var folders []MediaFolder = make([]MediaFolder, 0)
	tx := db.Model(&Media{}).Select("path", "name").Where("path", path).Where("directory", true)
	r := tx.Find(&folders)
	if r.Error != nil {
		return nil, r.Error
	}
	for i := range folders {
		folder := &folders[i]
		folder.Path = filepath.Join(folder.Path, folder.Name)
		tx := db.Model(&Media{}).Where("path", folder.Path)
		tx.Select("COUNT(*)").Where("directory", true).Find(&folder.FoldersCount)
		tx = db.Model(&Media{}).Where("path", folder.Path)
		tx.Select("COUNT(*)").Where("directory", false).Find(&folder.FilesCount)
	}
	return folders, r.Error
}
