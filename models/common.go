package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/restsend/carrot"
	"gorm.io/gorm"
)

type ContentIcon carrot.AdminIcon

func (s ContentIcon) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *ContentIcon) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), &s)
}

type StringArray []string

func (s StringArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringArray) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), &s)
}

type BaseContent struct {
	UpdatedAt   time.Time    `json:"updatedAt" gorm:"index"`
	CreatedAt   time.Time    `json:"createdAt" gorm:"index"`
	Thumbnail   string       `json:"thumbnail,omitempty" gorm:"size:500"`
	Tags        string       `json:"tags,omitempty" gorm:"size:200;index"`
	Title       string       `json:"title,omitempty" gorm:"size:200"`
	Alt         string       `json:"alt,omitempty"`
	Description string       `json:"description,omitempty"`
	Keywords    string       `json:"keywords,omitempty"`
	CreatorID   uint         `json:"-"`
	Creator     carrot.User  `json:"-"`
	Author      string       `json:"author" gorm:"size:64"`
	Published   bool         `json:"published"`
	PublishedAt sql.NullTime `json:"publishedAt"`
	ContentType string       `json:"contentType" gorm:"size:32"`
	Remark      string       `json:"remark"`
}

type SummaryResult struct {
	SiteCount     int64            `json:"sites"`
	PageCount     int64            `json:"pages"`
	PostCount     int64            `json:"posts"`
	CategoryCount int64            `json:"categories"`
	MediaCount    int64            `json:"media"`
	LatestPosts   []*RenderContent `json:"latestPosts"`
	BuildTime     string           `json:"buildTime"`
	CanExport     bool             `json:"canExport"`
}

func GetSummary(db *gorm.DB) (result SummaryResult) {
	db.Model(&Site{}).Count(&result.SiteCount)
	db.Model(&Page{}).Count(&result.PageCount)
	db.Model(&Post{}).Count(&result.PostCount)
	db.Model(&Category{}).Count(&result.CategoryCount)
	db.Model(&Media{}).Where("directory", false).Count(&result.MediaCount)

	var latestPosts []Post
	db.Order("updated_at desc").Limit(20).Find(&latestPosts)

	for idx := range latestPosts {
		item := NewRenderContentFromPost(db, &latestPosts[idx], false)
		item.PostBody = ""
		result.LatestPosts = append(result.LatestPosts, item)
	}
	return result
}
