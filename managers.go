package restcontent

import (
	"embed"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/restsend/restcontent/models"
	"gorm.io/gorm"
)

//go:embed admin
var EmbedAdminAssets embed.FS

type Manager struct {
	db                  *gorm.DB
	GitCommit           string
	BuildTime           string
	exportAndImportJobs sync.Map
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db, exportAndImportJobs: sync.Map{}}
}

func Migration(db *gorm.DB) error {
	return carrot.MakeMigrates(db, []any{
		&models.Site{},
		&models.Page{},
		&models.Post{},
		&models.Media{},
		&models.PublishLog{},
		&models.Category{},
	})
}

func (m *Manager) Prepare(engine *gin.Engine, lw io.Writer) error {
	logConfig := gin.LoggerConfig{
		Output:    lw,
		Formatter: m.httpLoggerFormat,
	}
	engine.Use(gin.LoggerWithConfig(logConfig))

	carrot.CheckValue(m.db, carrot.KEY_SITE_LOGO_URL, "/static/img/logo.svg")
	carrot.CheckValue(m.db, models.KEY_CMS_GUEST_ACCESS_API, "true")
	carrot.CheckValue(m.db, carrot.KEY_ADMIN_DASHBOARD, "./dashboard.html")
	carrot.CheckValue(m.db, models.KEY_CMS_UPLOAD_DIR, "./data/uploads/")
	carrot.CheckValue(m.db, models.KEY_CMS_MEDIA_PREFIX, "/media/")
	carrot.CheckValue(m.db, models.KEY_CMS_MEDIA_HOST, "")
	carrot.CheckValue(m.db, models.KEY_CMS_API_HOST, "")
	carrot.CheckValue(m.db, models.KEY_CMS_RELATION_COUNT, "3")
	carrot.CheckValue(m.db, models.KEY_CMS_SUGGESTION_COUNT, "3")

	if err := carrot.InitCarrot(m.db, engine); err != nil {
		return err
	}

	m.RegisterHandlers(engine)
	return nil
}

func (m *Manager) httpLoggerFormat(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	var userid string = "-"
	if user, ok := param.Keys[carrot.UserField]; ok && user != nil {
		userid = user.(*carrot.User).Email
	}

	return fmt.Sprintf("[HTTP] %v | %s |%s %3d %s| %s | %s | %15s |%s %-7s %s %#v\n%s",
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		userid,
		statusColor, param.StatusCode, resetColor,
		formatSizeHuman(float64(param.BodySize)),
		param.Latency.Round(time.Millisecond),
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

func formatSizeHuman(size float64) string {
	if size <= 0 {
		return "0 B"
	}
	if size < 1024 {
		return fmt.Sprintf("%.0f B", size)
	}
	size = size / 1024
	if size < 1024 {
		return fmt.Sprintf("%.1f KB", size)
	}
	size = size / 1024
	if size < 1024 {
		return fmt.Sprintf("%.1f MB", size)
	}
	size = size / 1024
	return fmt.Sprintf("%.1f GB", size)
}
