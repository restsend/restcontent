package restcontent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/restsend/restcontent/models"
)

func (m *Manager) RegisterHandlers(engine *gin.Engine) {
	admin := engine.Group("/admin", carrot.WithAdminAuth())
	handledObjects := carrot.BuildAdminObjects(admin, m.db, m.adminObjects())

	mediaPrefix := carrot.GetValue(m.db, models.KEY_CMS_MEDIA_PREFIX)
	if mediaPrefix == "" {
		mediaPrefix = "/media/"
	}
	media := engine.Group(mediaPrefix, m.AuthRequired)
	media.GET("/*filepath", m.handleMedia)

	admin.POST("/admin.json", func(ctx *gin.Context) {
		carrot.HandleAdminIndex(ctx, handledObjects, func(ctx *gin.Context, rc map[string]any) map[string]any {
			rc["dashboard"] = carrot.GetValue(m.db, carrot.KEY_ADMIN_DASHBOARD)
			rc["api_host"] = carrot.GetValue(m.db, models.KEY_CMS_API_HOST)

			rc["media_prefix"] = mediaPrefix
			rc["media_host"] = carrot.GetValue(m.db, models.KEY_CMS_MEDIA_HOST)
			rc["build_time"] = m.BuildTime
			return rc
		})
	})
	admin.POST("/summary", m.handleAdminSummary)

	admin.StaticFS("/", carrot.NewCombineEmbedFS(
		carrot.HintAssetsRoot("admin"),                                // dev assets
		carrot.EmbedFS{EmbedRoot: "admin", Embedfs: EmbedAdminAssets}, // restcontent's embed admin
		carrot.EmbedFS{EmbedRoot: "admin", Embedfs: carrot.EmbedAdminAssets}))

	admin.POST("/export/start", m.superAccessCheck, m.handleExportStart)
	admin.POST("/export/poll", m.superAccessCheck, m.handleExportPoll)

	admin.POST("/import/upload", m.superAccessCheck, m.handleImportUpload)
	admin.POST("/import/start", m.superAccessCheck, m.handleImportStart)
	admin.POST("/import/poll", m.superAccessCheck, m.handleImportPoll)

	prefix := carrot.GetEnv(models.ENV_CMS_API_PREFIX)
	if prefix == "" {
		prefix = "/api"
	}
	routes := engine.Group(prefix, m.AuthRequired)
	objs := []carrot.WebObject{
		{
			Model:        &models.Site{},
			AllowMethods: carrot.GET | carrot.QUERY,
			Name:         "site",
			Editables:    []string{"Domain", "Name", "Preview", "Disallow"},
			Filterables:  []string{},
			Orderables:   []string{},
			Searchables:  []string{"Domain", "Name"},
		},
		{
			Model:        &models.Category{},
			AllowMethods: carrot.GET | carrot.QUERY,
			Name:         "category",
			Editables:    []string{"UUID", "SiteID", "Name", "Items"},
			Filterables:  []string{},
			Orderables:   []string{},
			Searchables:  []string{"UUID", "Name", "Items"},
		},

		{
			Model:        &models.Page{},
			AllowMethods: carrot.GET | carrot.QUERY,
			Name:         "page",
			Filterables:  []string{"SiteID", "CategoryID", "CategoryPath", "Tags", "IsDraft", "Published", "ContentType"},
			Searchables:  []string{"Title", "Description", "Body"},
			GetDB:        m.getPostOrPageDB,
			BeforeRender: m.beforeRenderPage,
		},
		{
			Model:             &models.Post{},
			AllowMethods:      carrot.GET | carrot.QUERY,
			Name:              "post",
			Filterables:       []string{"SiteID", "CategoryID", "CategoryPath", "Tags", "IsDraft", "Published", "ContentType"},
			Searchables:       []string{"Title", "Description", "Body"},
			GetDB:             m.getPostOrPageDB,
			BeforeRender:      m.beforeRenderPost,
			BeforeQueryRender: m.beforeQueryRenderPost,
		},
	}
	carrot.RegisterObjects(routes, objs)
}

func (m *Manager) AuthRequired(c *gin.Context) {
	if carrot.CurrentUser(c) != nil {
		c.Next()
		return
	}

	guestAccess := carrot.GetBoolValue(m.db, models.KEY_CMS_GUEST_ACCESS_API)
	if guestAccess {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions:
			c.Next()
			return
		}
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		carrot.AbortWithJSONError(c, http.StatusUnauthorized, models.ErrUnauthorized)
		return
	}
	// split bearer
	token = token[len("Bearer "):]
	user, err := carrot.DecodeHashToken(m.db, token, false)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusUnauthorized, err)
		return
	}
	c.Set(carrot.UserField, user)
	c.Next()
}
