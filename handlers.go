package restcontent

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/restsend/restcontent/models"
	"gorm.io/gorm"
)

func readIcon(name string) *carrot.AdminIcon {
	data, err := EmbedAdminAssets.ReadFile(filepath.Join("admin", name))
	if err != nil {
		carrot.Warning("Read icon failed:", name, err)
		return nil
	}
	return &carrot.AdminIcon{SVG: string(data)}
}

func (m *Manager) adminObjects() []carrot.AdminObject {

	vals := []carrot.AdminObject{
		{
			Model: &models.Site{},
			Group: "Contents",
			Name:  "Site",
			Shows: []string{"Domain", "Name", "Preview", "Disallow", "UpdatedAt", "CreatedAt"},
			Orders: []carrot.Order{
				{
					Name: "UpdatedAt",
					Op:   carrot.OrderOpDesc,
				},
			},
			Editables:   []string{"Domain", "Name", "Preview", "Disallow"},
			Filterables: []string{"Disallow"},
			Orderables:  []string{},
			Searchables: []string{"Domain", "Name", "Preview"},
			Requireds:   []string{"Domain"},
			Icon:        readIcon("./icon/desktop.svg"),
			Scripts: []carrot.AdminScript{
				{Src: "./js/cms_site.js", Onload: true},
			},
		},
		{
			Model:       &models.Category{},
			Group:       "Contents",
			Name:        "Category",
			Desc:        "The category of articles and pages can be multi-level",
			Shows:       []string{"Name", "UUID", "Site", "Items"},
			Editables:   []string{"Name", "UUID", "Site", "Items"},
			Orderables:  []string{},
			Searchables: []string{"UUID", "Site", "Items", "Name"},
			Requireds:   []string{"UUID", "Site", "Items", "Name"},
			Icon:        readIcon("./icon/swatch.svg"),
			Attributes:  map[string]carrot.AdminAttribute{"Items": {Widget: "category-item"}},
			Scripts: []carrot.AdminScript{
				{Src: "./js/cms_widget.js"},
				{Src: "./js/cms_category.js", Onload: true},
			},
			Actions: []carrot.AdminAction{
				{
					WithoutObject: true,
					Path:          "query_with_count",
					Name:          "Query with item count",
					Handler:       m.handleQueryCategoryWithCount,
				},
			},
		},
		m.getPageObject(),
		m.getPostObject(),
		m.getMediaObject(),
		{
			Model:     &models.PublishLog{},
			Invisible: true,
			Group:     "Contents",
			Name:      "PublishLog",
			Desc:      "Post and Page publish log",
			Shows:     []string{"ID", "Author", "ContentID", "CreatedAt"},
			Orders: []carrot.Order{
				{
					Name: "CreatedAt",
					Op:   carrot.OrderOpDesc,
				},
			},
			Editables: []string{"ID", "Author", "ContentID", "Body"},
		},
	}
	settings := carrot.GetCarrotAdminObjects()
	vals = append(vals, settings...)
	carrot.Warning("Admin objects count:", len(vals))
	return vals
}

func (m *Manager) handleQueryCategoryWithCount(db *gorm.DB, c *gin.Context, obj any) (any, error) {
	siteId := c.Query("site_id")
	current := strings.ToLower(c.Query("current"))
	return models.QueryCategoryWithCount(db, siteId, current)
}

func (m *Manager) handleAdminSummary(c *gin.Context) {
	result := models.GetSummary(m.db)
	result.BuildTime = m.BuildTime
	result.CanExport = carrot.CurrentUser(c).IsSuperUser
	c.JSON(http.StatusOK, result)
}
