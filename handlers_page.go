package restcontent

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/restsend/restcontent/models"
	"gorm.io/gorm"
)

var enabledPageContentTypes = []carrot.AdminSelectOption{
	{Value: models.ContentTypeJson, Label: "JSON"},
	{Value: models.ContentTypeHtml, Label: "HTML"},
	{Value: models.ContentTypeMarkdown, Label: "Markdown"},
}

func (m *Manager) getPageObject() carrot.AdminObject {

	return carrot.AdminObject{
		Model:       &models.Page{},
		Group:       "Contents",
		Name:        "Page",
		Desc:        "The page data of the website can only be in JSON/YAML format",
		Shows:       []string{"ID", "Site", "Title", "Author", "IsDraft", "Published", "PublishedAt", "CategoryID", "Tags", "CreatedAt"},
		Editables:   []string{"ID", "Site", "CategoryID", "CategoryPath", "Author", "IsDraft", "Draft", "Published", "PublishedAt", "ContentType", "Thumbnail", "Tags", "Title", "Alt", "Description", "Keywords", "Draft", "Remark"},
		Filterables: []string{"Site", "CategoryID", "Tags", "Published", "UpdatedAt"},
		Orderables:  []string{"UpdatedAt", "PublishedAt"},
		Searchables: []string{"ID", "Tags", "Title", "Alt", "Description", "Keywords", "Body"},
		Requireds:   []string{"ID", "Site", "CategoryID", "ContentType", "Body"},
		Icon:        readIcon("./icon/piece.svg"),
		Styles: []string{
			"./css/jsoneditor-9.10.2.min.css",
		},
		Scripts: []carrot.AdminScript{
			{Src: "./js/cms_widget.js"},
			{Src: "./js/jsoneditor-9.10.2.min.js"},
			{Src: "./js/cms_page.js", Onload: true}},
		Attributes: map[string]carrot.AdminAttribute{
			"ContentType": {Choices: enabledPageContentTypes, Default: models.ContentTypeJson},
			"Draft":       {Default: "{}"},
			"IsDraft":     {Widget: "is-draft"},
			"Published":   {Widget: "is-published"},
			"Tags":        {Widget: "tags", FilterWidget: "tags"},
			"CategoryID":  {Widget: "category-id-and-path", FilterWidget: "category-id-and-path"},
			"ID":          {Help: "ID must be unique,recommend use page url eg: about-us"},
		},
		EditPage: "./edit_page.html",
		Orders: []carrot.Order{
			{
				Name: "UpdatedAt",
				Op:   carrot.OrderOpDesc,
			},
		},
		Actions: []carrot.AdminAction{
			{
				WithoutObject: true,
				Path:          "save_draft",
				Name:          "Safe Draft",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleSaveDraft(db, c, obj)
				},
			},
			{
				Path: "make_publish",
				Name: "Make Publish",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleMakePagePublish(db, c, obj, true)
				},
			},
			{
				Path: "make_un_publish",
				Name: "Make UnPublish",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleMakePagePublish(db, c, obj, false)
				},
			},
			{
				WithoutObject: true,
				Path:          "tags",
				Name:          "Query All Tags",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleQueryTags(db, c, obj, "pages")
				},
			},
		},
		BeforeCreate: func(db *gorm.DB, ctx *gin.Context, vptr any) error {
			page := vptr.(*models.Page)
			page.ContentType = models.ContentTypeJson
			page.Creator = *carrot.CurrentUser(ctx)
			page.IsDraft = true
			return nil
		},
		BeforeUpdate: func(db *gorm.DB, ctx *gin.Context, vptr any, vals map[string]any) error {
			page := vptr.(*models.Page)
			page.IsDraft = true
			if _, ok := vals["published"]; ok {
				page.Published = vals["published"].(bool)
				if page.Published {
					page.Body = page.Draft
					page.IsDraft = false
				}
			}
			return nil
		},
	}
}

func (m *Manager) getPostObject() carrot.AdminObject {
	return carrot.AdminObject{
		Model:       &models.Post{},
		Group:       "Contents",
		Name:        "Post",
		Desc:        "Website articles or blogs, support HTML and Markdown formats",
		Shows:       []string{"ID", "Site", "Title", "Author", "CategoryID", "Tags", "IsDraft", "Published", "PublishedAt", "CreatedAt"},
		Editables:   []string{"ID", "Site", "CategoryID", "CategoryPath", "Author", "IsDraft", "Draft", "Published", "PublishedAt", "ContentType", "Thumbnail", "Tags", "Title", "Alt", "Description", "Keywords", "Draft", "Remark"},
		Filterables: []string{"Site", "CategoryID", "Tags", "Published", "UpdatedAt"},
		Orderables:  []string{"UpdatedAt", "PublishedAt"},
		Searchables: []string{"ID", "Tags", "Title", "Alt", "Description", "Keywords", "Body"},
		Requireds:   []string{"ID", "Site", "CategoryID", "ContentType", "Body"},
		Icon:        readIcon("./icon/newspaper.svg"),
		Styles: []string{
			"./css/easymde.min.css",
			"./css/jodit.min.css",
		},
		Scripts: []carrot.AdminScript{
			{Src: "./js/cms_widget.js"},
			{Src: "./js/easymde.min.js"},
			{Src: "./js/jodit.min.js"},
			{Src: "./js/cms_page.js", Onload: true}},
		Attributes: map[string]carrot.AdminAttribute{
			"ContentType": {Choices: enabledPageContentTypes, Default: models.ContentTypeHtml},
			"Draft":       {Default: "Your content ..."},
			"IsDraft":     {Widget: "is-draft"},
			"Published":   {Widget: "is-published"},
			"Tags":        {Widget: "tags", FilterWidget: "tags"},
			"CategoryID":  {Widget: "category-id-and-path", FilterWidget: "category-id-and-path"},
			"ID":          {Help: "ID must be unique,recommend use title slug eg: hello-world-2023"},
		},
		EditPage: "./edit_page.html",
		Orders: []carrot.Order{
			{
				Name: "UpdatedAt",
				Op:   carrot.OrderOpDesc,
			},
		},
		Actions: []carrot.AdminAction{
			{
				WithoutObject: true,
				Path:          "save_draft",
				Name:          "Safe Draft",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleSaveDraft(db, c, obj)
				},
			},
			{
				Path: "make_publish",
				Name: "Make Publish",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleMakePagePublish(db, c, obj, true)
				},
			},
			{
				Path: "make_un_publish",
				Name: "Make UnPublish",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleMakePagePublish(db, c, obj, false)
				},
			},
			{
				WithoutObject: true,
				Path:          "tags",
				Name:          "Query All Tags",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleQueryTags(db, c, obj, "posts")
				},
			},
		},
		BeforeCreate: func(db *gorm.DB, ctx *gin.Context, vptr any) error {
			post := vptr.(*models.Post)
			if post.ContentType == "" {
				post.ContentType = models.ContentTypeMarkdown
			}
			post.Creator = *carrot.CurrentUser(ctx)
			post.IsDraft = true
			return nil
		},
		BeforeUpdate: func(db *gorm.DB, ctx *gin.Context, vptr any, vals map[string]any) error {
			post := vptr.(*models.Post)
			post.IsDraft = true
			if _, ok := vals["published"]; ok {
				post.Published = vals["published"].(bool)
				if post.Published {
					post.Body = post.Draft
					post.IsDraft = false
				}
			}
			return nil
		},
	}
}

func (m *Manager) handleMakePagePublish(db *gorm.DB, c *gin.Context, obj any, publish bool) (any, error) {
	siteId := c.Query("site_id")
	id := c.Query("id")
	if err := models.MakePublish(db, siteId, id, obj, publish); err != nil {
		carrot.Warning("make publish failed:", siteId, id, publish, err)
		return false, err
	}
	return true, nil
}

func (m *Manager) handleSaveDraft(db *gorm.DB, c *gin.Context, obj any) (any, error) {
	siteId := c.Query("site_id")
	id := c.Query("id")

	var formData map[string]string
	if err := c.ShouldBind(&formData); err != nil {
		return nil, err
	}

	draft, ok := formData["draft"]
	if !ok {
		return nil, models.ErrDraftIsInvalid
	}

	if err := models.SafeDraft(db, siteId, id, obj, draft); err != nil {
		carrot.Warning("safe draft failed:", siteId, id, err)
		return false, err
	}
	return true, nil
}

func (m *Manager) handleQueryTags(db *gorm.DB, c *gin.Context, obj any, tableName string) (any, error) {
	return models.QueryTags(db.Table(tableName))
}

func (m *Manager) beforeRenderPage(db *gorm.DB, ctx *gin.Context, vptr any) (any, error) {
	draft, _ := strconv.ParseBool(ctx.Query("draft"))
	result := vptr.(*models.Page)
	if !draft && !result.Published {
		carrot.AbortWithJSONError(ctx, http.StatusTooEarly, models.ErrPageIsNotPublish)
		return nil, models.ErrPageIsNotPublish
	}
	if draft {
		result.Body = result.Draft
	}
	return models.NewRenderContentFromPage(m.db, result), nil
}

func (m *Manager) getPostOrPageDB(ctx *gin.Context, isCreate bool) *gorm.DB {
	if isCreate {
		return m.db
	}
	draft, _ := strconv.ParseBool(ctx.Query("draft"))
	if draft {
		return m.db
	}
	// single get not need published
	if ctx.Request.Method == http.MethodGet {
		return m.db
	}
	// query must be published
	return m.db.Where("published", true)
}

func (m *Manager) beforeRenderPost(db *gorm.DB, ctx *gin.Context, vptr any) (any, error) {
	draft, _ := strconv.ParseBool(ctx.Query("draft"))
	result := vptr.(*models.Post)
	if !draft && !result.Published {
		return nil, models.ErrPostIsNotPublish
	}
	if draft {
		result.Body = result.Draft
	}

	relations := true
	if ctx.Request.Method == http.MethodPost { // batch query
		relations = false
	}

	return models.NewRenderContentFromPost(m.db, result, relations), nil
}

func (m *Manager) beforeQueryRenderPost(db *gorm.DB, ctx *gin.Context, queryResult *carrot.QueryResult) (any, error) {
	if len(queryResult.Items) <= 0 {
		return nil, nil
	}
	firstItem, ok := queryResult.Items[0].(*models.RenderContent)
	if !ok {
		return nil, nil
	}

	siteId := firstItem.SiteID
	categoryId := ""
	categoryPath := ""
	if firstItem.Category != nil {
		categoryId = firstItem.Category.UUID
		categoryPath = firstItem.Category.Path
	}

	r := &models.ContentQueryResult{
		QueryResult: queryResult,
	}

	relationCount := carrot.GetIntValue(m.db, models.KEY_CMS_RELATION_COUNT, 3)
	suggestionCount := carrot.GetIntValue(m.db, models.KEY_CMS_SUGGESTION_COUNT, 3)

	r.Suggestions, _ = models.GetSuggestions(m.db, siteId, categoryId, categoryPath, "", relationCount)
	r.Relations, _ = models.GetRelations(m.db, siteId, categoryId, categoryPath, "", suggestionCount)

	return r, nil
}
