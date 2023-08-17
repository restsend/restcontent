package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/restsend/restcontent"
	"gorm.io/gorm"
)

var GitCommit string
var BuildTime string
var setupDoneFlag string = ".restcontent_setup_done"

type SetupDBForm struct {
	Driver   string `json:"dbDriver"`
	Host     string `json:"dbHost"`
	Port     string `json:"dbPort"`
	Name     string `json:"dbName"`
	Filename string `json:"dbFilename"`
	Charset  string `json:"dbCharset"`
	User     string `json:"dbUser"`
	Password string `json:"dbPassword"`
}

type SetupSuperUserForm struct {
	DBConfig SetupDBForm `json:"dbConfig" binding:"required"`
	Username string      `json:"superUsername" binding:"required"`
	Password string      `json:"superPassword" binding:"required"`
}

type SetupSuperEnvForm struct {
	DBConfig     SetupDBForm `json:"dbConfig" binding:"required"`
	Salt         string      `json:"salt" binding:"required"`
	CookieSecret string      `json:"cookieSecret" binding:"required"`
	LogFile      string      `json:"logFile"`
}

func (f *SetupDBForm) DSN() string {
	if f.Driver == "sqlite" {
		return fmt.Sprintf("file:%s", f.Filename)
	}

	pwd := f.Password
	if pwd == "" {
		pwd = "''"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		f.User, pwd, f.Host, f.Port, f.Name, f.Charset)
}

func runSetupMode(addr string) {
	carrot.Warning("Run setup mode")
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	carrot.Warning("Please visit http://", addr, "/setup to complete install")

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	as := carrot.NewStaticAssets()
	as.InitStaticAssets(r)
	r.HTMLRender = as

	srv := &http.Server{Handler: r.Handler()}

	r.GET("/admin", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/setup/")
	})

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/setup/")
	})

	r.GET("/setup", func(c *gin.Context) {
		osVersion := fmt.Sprintf("%s-%s (%s)", runtime.GOOS, runtime.GOARCH, runtime.Version())
		// current working directory
		cwd, _ := os.Getwd()
		ctx := map[string]any{
			"buildTime":    BuildTime,
			"gitCommit":    GitCommit,
			"osVersion":    osVersion,
			"cwd":          cwd,
			"enableSqlite": enableSqlite,
		}
		c.HTML(200, "setup.html", ctx)
	})

	r.POST("/setup/ping_database", func(ctx *gin.Context) {
		var form SetupDBForm
		if err := ctx.ShouldBindJSON(&form); err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}

		var err error
		var db *gorm.DB
		defer func() {
			if db != nil {
				h, _ := db.DB()
				h.Close()
			}
		}()

		carrot.Warning("DSN", form.DSN())
		db, err = carrot.InitDatabase(os.Stdout, form.Driver, form.DSN())
		if err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}
		ctx.JSON(200, true)
	})

	r.POST("/setup/migrate_database", func(ctx *gin.Context) {
		var form SetupDBForm
		if err := ctx.ShouldBindJSON(&form); err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}

		var err error
		var db *gorm.DB
		defer func() {
			if db != nil {
				h, _ := db.DB()
				h.Close()
			}
		}()
		db, err = carrot.InitDatabase(os.Stdout, form.Driver, form.DSN())
		if err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}
		err = carrot.InitMigrate(db)
		if err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}
		err = restcontent.Migration(db)
		if err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}
		ctx.JSON(200, true)
	})

	r.POST("/setup/write_env", func(ctx *gin.Context) {
		var form SetupSuperEnvForm
		if err := ctx.ShouldBindJSON(&form); err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}
		envFile := ".env"

		lines := []string{
			fmt.Sprintf("%s=%s", carrot.ENV_SALT, form.Salt),
			fmt.Sprintf("%s=%s", carrot.ENV_SESSION_SECRET, form.CookieSecret),
			fmt.Sprintf("LOG_FILE=%s", form.LogFile),
			fmt.Sprintf("DSN=%s", form.DBConfig.DSN()),
			fmt.Sprintf("DB_DRIVER=%s", form.DBConfig.Driver),
		}
		data := strings.Join(lines, "\n") + "\n"
		if _, err := os.Stat(envFile); err == nil {
			fileData, _ := os.ReadFile(envFile)
			if fileData != nil {
				data = string(fileData) + data
			}
		}
		os.WriteFile(envFile, []byte(data), 0644)
		ctx.JSON(200, true)
	})

	r.POST("/setup/create_superuser", func(ctx *gin.Context) {
		var form SetupSuperUserForm
		if err := ctx.ShouldBindJSON(&form); err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}

		var err error
		var db *gorm.DB
		defer func() {
			if db != nil {
				h, _ := db.DB()
				h.Close()
			}
		}()

		db, err = carrot.InitDatabase(os.Stdout, form.DBConfig.Driver, form.DBConfig.DSN())
		if err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}
		err = carrot.InitMigrate(db)
		if err != nil {
			ctx.Data(400, "text/plain", []byte(err.Error()))
			return
		}

		u, err := carrot.GetUserByEmail(db, form.Username)
		if err == nil && u != nil {
			carrot.SetPassword(db, u, form.Password)
			carrot.Warning("Update super with new password")
		} else {
			u, err = carrot.CreateUser(db, form.Username, form.Password)
			if err != nil {
				panic(err)
			}
		}
		u.IsStaff = true
		u.Activated = true
		u.Enabled = true
		u.IsSuperUser = true
		db.Save(u)
		carrot.Warning("Create super user:", form.Username)
		ctx.JSON(200, true)
	})

	r.POST("/setup/restart", func(ctx *gin.Context) {
		ctx.JSON(200, true)
		os.WriteFile(setupDoneFlag, []byte("done"), 0644)
		time.AfterFunc(500*time.Millisecond, func() {
			carrot.Warning("Restarting...")
			srv.Shutdown(context.Background())
		})
	})
	srv.Serve(ln)
}
