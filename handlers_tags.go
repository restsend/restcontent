package restcontent

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/restsend/restcontent/models"
)

func (m *Manager) handleGetTags(c *gin.Context) {
	contentType := c.Param("content_type")
	var form models.TagsForm
	if err := c.BindJSON(&form); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	tags, err := models.GetTagsByCategory(m.db, contentType, &form)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, tags)
}
