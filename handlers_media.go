package restcontent

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/restsend/restcontent/models"
	"gorm.io/gorm"
)

func (m *Manager) getMediaObject() carrot.AdminObject {
	return carrot.AdminObject{
		Model:       &models.Media{},
		Group:       "Contents",
		Name:        "Media",
		Desc:        "All kinds of media files, such as images, videos, audios, etc.",
		Shows:       []string{"Name", "ContentType", "Author", "Published", "Size", "Dimensions", "UpdatedAt"},
		Editables:   []string{"External", "PublicUrl", "Author", "Published", "PublishedAt", "Title", "Alt", "Description", "Keywords", "ContentType", "Size", "Path", "Name", "Dimensions", "StorePath", "UpdatedAt", "Ext", "Size", "StorePath", "Remark"},
		Filterables: []string{"Published", "UpdatedAt", "ContentType", "External"},
		Orderables:  []string{"UpdatedAt", "PublishedAt", "Size"},
		Searchables: []string{"Title", "Alt", "Description", "Keywords", "Path", "Path", "Name", "StorePath"},
		Requireds:   []string{"ContentType", "Size", "Path", "Name", "Dimensions", "StorePath"},
		Icon:        readIcon("./icon/image.svg"),
		Attributes: map[string]carrot.AdminAttribute{
			"ContentType": {Choices: models.ContentTypes},
			"Size":        {Widget: "humanize-size"},
			"Site":        {SingleChoice: true},
		},
		Scripts: []carrot.AdminScript{
			{Src: "./js/cms_widget.js"},
			{Src: "./js/cms_media.js", Onload: true},
		},
		EditPage: "./edit_media.html",
		Orders: []carrot.Order{
			{
				Name: "UpdatedAt",
				Op:   carrot.OrderOpAsc,
			},
		},
		BeforeRender: func(db *gorm.DB, ctx *gin.Context, vptr any) (any, error) {
			media := vptr.(*models.Media)
			mediaHost := carrot.GetValue(db, models.KEY_CMS_MEDIA_HOST)
			mediaPrefix := carrot.GetValue(db, models.KEY_CMS_MEDIA_PREFIX)
			media.BuildPublicUrls(mediaHost, mediaPrefix)
			return vptr, nil
		},
		BeforeCreate: func(db *gorm.DB, ctx *gin.Context, vptr any) error {
			media := vptr.(*models.Media)
			media.Creator = *carrot.CurrentUser(ctx)
			return nil
		},
		BeforeDelete: func(db *gorm.DB, ctx *gin.Context, vptr any) error {
			media := vptr.(*models.Media)
			if err := models.RemoveFile(db, media.Path, media.Name); err != nil {
				carrot.Warning("Delete file failed: ", media.StorePath, err)
			}
			return nil
		},
		Actions: []carrot.AdminAction{
			{
				Path: "make_publish",
				Name: "Make Publish",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleMakeMediaPublish(db, c, obj, true)
				},
			},
			{
				Path: "make_un_publish",
				Name: "Make UnPublish",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (any, error) {
					return m.handleMakeMediaPublish(db, c, obj, false)
				},
			},
			{
				WithoutObject: true,
				Path:          "folders",
				Name:          "Folders",
				Handler:       m.handleListFolders,
			},
			{
				WithoutObject: true,
				Path:          "new_folder",
				Name:          "New Folder",
				Handler:       m.handleNewFolder,
			},
			{
				WithoutObject: true,
				Path:          "upload",
				Name:          "Upload",
				Handler:       m.handleUpload,
			},
			{
				WithoutObject: true,
				Path:          "remove_dir",
				Name:          "Remove directory",
				Handler:       m.handleRemoveDirectory,
			},
		},
	}
}

func (m *Manager) handleListFolders(db *gorm.DB, c *gin.Context, obj any) (any, error) {
	path := c.Query("path")
	return models.ListFolders(db, path)
}

func (m *Manager) handleNewFolder(db *gorm.DB, c *gin.Context, obj any) (any, error) {
	path := c.Query("path")
	name := c.Query("name")
	user := carrot.CurrentUser(c)
	return models.CreateFolder(db, path, name, user)
}

func (m *Manager) handleMakeMediaPublish(db *gorm.DB, c *gin.Context, obj any, publish bool) (any, error) {
	siteId := c.Query("site_id")
	path := c.Query("path")
	name := c.Query("name")

	if err := models.MakeMediaPublish(db, siteId, path, name, obj, publish); err != nil {
		carrot.Warning("Make publish failed:", siteId, path, name, publish, err)
		return false, err
	}
	return true, nil
}

func (m *Manager) handleMedia(c *gin.Context) {
	fullPath := c.Param("filepath")
	path, name := filepath.Split(fullPath)
	img, err := models.GetMedia(m.db, path, name)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusNotFound, err)
		return
	}

	if img.External {
		c.Redirect(http.StatusFound, img.StorePath)
		return
	}

	uploadDir := carrot.GetValue(m.db, models.KEY_CMS_UPLOAD_DIR)
	filepath := filepath.Join(uploadDir, img.StorePath)
	http.ServeFile(c.Writer, c.Request, filepath)
}

func (m *Manager) handleRemoveDirectory(db *gorm.DB, c *gin.Context, obj any) (any, error) {
	path := c.Query("path")

	parent, err := models.RemoveDirectory(db, path)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return nil, err
	}
	return parent, nil
}

func (m *Manager) handleUpload(db *gorm.DB, c *gin.Context, obj any) (any, error) {
	created := c.Query("created")
	path := c.Query("path")
	name := c.Query("name")

	file, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}

	mFile, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer mFile.Close()

	if path == "" {
		path = "/"
	}
	if name == "" {
		name = file.Filename
	}
	r, err := models.UploadFile(db, path, name, mFile)
	if err != nil {
		return nil, err
	}

	var media models.Media

	user := carrot.CurrentUser(c)
	media.Name = r.Name
	media.Path = r.Path
	media.External = r.External
	media.StorePath = r.StorePath
	media.Size = r.Size
	media.ContentType = r.ContentType
	media.Dimensions = r.Dimensions
	media.Directory = false
	media.Ext = r.Ext
	media.ContentType = r.ContentType
	media.Published = true

	if user != nil {
		media.Creator = *user
		media.CreatorID = user.ID
	}

	if created != "" {
		result := db.Create(&media)
		if result.Error != nil {
			return nil, result.Error
		}
	}

	mediaHost := carrot.GetValue(m.db, models.KEY_CMS_MEDIA_HOST)
	mediaPrefix := carrot.GetValue(m.db, models.KEY_CMS_MEDIA_PREFIX)
	media.BuildPublicUrls(mediaHost, mediaPrefix)

	r.PublicUrl = media.PublicUrl
	r.Thumbnail = media.Thumbnail

	return r, nil
}
