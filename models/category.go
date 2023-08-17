package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type CategoryItem struct {
	Path     string        `json:"path"`
	Name     string        `json:"name"`
	Icon     *ContentIcon  `json:"icon,omitempty"`
	Children CategoryItems `json:"children,omitempty"`
	Count    int           `json:"count" gorm:"-"`
}

type Category struct {
	SiteID string        `json:"siteId" gorm:"uniqueIndex:,composite:_site_uuid"`
	Site   Site          `json:"-"`
	UUID   string        `json:"uuid" gorm:"size:12;uniqueIndex:,composite:_site_uuid"`
	Name   string        `json:"name" gorm:"size:200"`
	Items  CategoryItems `json:"items,omitempty"`
	Count  int           `json:"count" gorm:"-"`
}

type RenderCategory struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Path     string `json:"path,omitempty"`
	PathName string `json:"pathName,omitempty"`
}

func (s CategoryItem) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *CategoryItem) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), &s)
}

type CategoryItems []CategoryItem

func (s CategoryItems) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *CategoryItems) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), &s)
}

func (category *Category) findItem(path string, items CategoryItems) *CategoryItem {
	for _, item := range items {
		if item.Path == path {
			return &item
		}

		if item.Children != nil {
			if found := category.findItem(path, item.Children); found != nil {
				return found
			}
		}
	}
	return nil
}

func (category *Category) FindItem(path string) *CategoryItem {
	if path == "" {
		return nil
	}
	return category.findItem(path, category.Items)
}

func QueryCategoryWithCount(db *gorm.DB, siteId, contentObject string) ([]Category, error) {
	var tx *gorm.DB
	switch contentObject {
	case "post":
		tx = db.Model(&Post{}).Where("site_id", siteId)
	case "page":
		tx = db.Model(&Page{}).Where("site_id", siteId)
	default:
		return nil, fmt.Errorf("invalid content object: %s", contentObject)
	}

	var vals []Category
	r := db.Model(&Category{}).Where("site_id", siteId).Find(&vals)
	if r.Error != nil {
		return nil, r.Error
	}
	for i := range vals {
		val := &vals[i]
		tx := tx.Where("category_id", val.UUID)
		var count int64
		tx.Count(&count)
		val.Count = int(count)
	}
	return vals, r.Error
}

func NewRenderCategory(db *gorm.DB, categoryID, categoryPath string) *RenderCategory {
	var category Category
	r := db.Model(&Category{}).Where("uuid", categoryID).First(&category)
	if r.Error != nil {
		return nil
	}

	selected := category.FindItem(categoryPath)

	obj := &RenderCategory{
		UUID: category.UUID,
		Name: category.Name,
	}
	if selected != nil {
		obj.Path = selected.Path
		obj.PathName = selected.Name
	}
	return obj
}
