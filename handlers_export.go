package restcontent

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/restsend/carrot"
	"github.com/restsend/restcontent/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type rowExportHandle func(out *zip.Writer, modelObj any) (int64, bool, error)
type rowImportHandle func(in *zip.Reader, modelObj any) (bool, error)

type ExportResult struct {
	Reason       string `json:"reason,omitempty"`
	Status       string `json:"status,omitempty"`
	DownloadLink string `json:"downloadLink,omitempty"`
	DownloadSize int64  `json:"downloadSize,omitempty"`
}

type ExportJob struct {
	m         *Manager
	user      *carrot.User
	result    ExportResult
	mutex     sync.Mutex
	key       string
	Options   []string `json:"options" binding:"required"`
	MediaHost string
	From      string
}

type ExportOption struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	Size  int64  `json:"size"`
}

type ExportMeta struct {
	BuildTime   string         `json:"buildTime"`
	Options     []ExportOption `json:"options"`
	From        string         `json:"from"`
	MediaHost   string         `json:"mediaHost"`
	MediaPrefix string         `json:"mediaPrefix"`
	ExportTime  time.Time      `json:"exportTime"`
	Author      string         `json:"author,omitempty"`
	Key         string         `json:"key,omitempty"`
}

type ImportJob struct {
	result      ExportResult
	mutex       sync.Mutex
	user        *carrot.User
	m           *Manager
	Meta        ExportMeta
	key         string
	TmpFileName string
}

type StartImportForm struct {
	Key     string   `json:"key" binding:"required"`
	Options []string `json:"options" binding:"required"`
}

func (job *ImportJob) Start(options []string) {
	job.result.Status = "pending"

	carrot.Warning("Import job start: ", job.key, job.TmpFileName, options)
	zipFile, err := os.Open(job.TmpFileName)
	if err != nil {
		job.result.Status = "error"
		job.result.Reason = "open tmp file: " + err.Error()
		return
	}

	var fileSize int64
	if fi, err := zipFile.Stat(); err == nil {
		fileSize = fi.Size()
	}

	zipReader, err := zip.NewReader(zipFile, fileSize)
	if err != nil {
		job.result.Status = "error"
		job.result.Reason = fmt.Sprintf("zip file error: %v", err)
		return
	}

	go func() {
		tx := job.m.db.Begin()
		defer func() {
			if tx != nil {
				tx.Rollback()
				tx = nil
			}

			zipFile.Close()
			if err := recover(); err != nil {
				job.mutex.Lock()
				job.result.Status = "error"
				job.result.Reason = fmt.Sprintf("recover: %v", err)
				job.mutex.Unlock()
				carrot.Warning("Import job crash:", err)
			}
		}()

		defer time.AfterFunc(5*time.Second, func() {
			job.m.exportAndImportJobs.Delete(job.key)
		})

		for _, opt := range options {
			switch opt {
			case "users", "sites", "categories", "pages", "posts", "media":
			default:
				continue
			}

			err := job.Import(tx, zipReader, opt)
			if err != nil {
				job.mutex.Lock()
				job.result.Status = "error"
				job.result.Reason = fmt.Sprintf("[%s] import %s", opt, err.Error())
				job.mutex.Unlock()
				return
			}
		}
		tx.Commit()
		tx = nil
		job.mutex.Lock()
		job.result.Status = "done"
		job.mutex.Unlock()
	}()
}

func (job *ImportJob) GetResult() ExportResult {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	r := job.result
	return r
}

func (job *ImportJob) importTable(tx *gorm.DB, zipReader *zip.Reader, opt string, rowHandle rowImportHandle) error {
	f, err := zipReader.Open(fmt.Sprintf("%s.json", opt))
	if err != nil {
		return nil
	}

	data := bytes.NewBuffer(nil)
	io.Copy(data, f)
	f.Close()

	var lines []map[string]any
	if err := json.Unmarshal(data.Bytes(), &lines); err != nil {
		return err
	}

	obj, err := getAdminObject(job.m.db, opt)
	if err != nil {
		return err
	}

	for _, line := range lines {
		modelElem := reflect.New(reflect.TypeOf(obj.Model).Elem())
		modelObj, err := obj.UnmarshalFrom(modelElem, nil, line)
		if err != nil {
			return fmt.Errorf("unmarshal failed: %v", err)
		}

		if rowHandle != nil {
			ok, err := rowHandle(zipReader, modelObj)
			if !ok {
				continue
			}
			if err != nil {
				carrot.Warning("Import row with handle failed: ", modelObj, err)
				continue
			}
		}

		if err := tx.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(modelObj).Error; err != nil {
			carrot.Warning("Import row failed: ", modelObj, err)
			return fmt.Errorf("create failed: %v", err)
		}
	}
	return nil
}

func (job *ImportJob) Import(tx *gorm.DB, zipReader *zip.Reader, opt string) error {
	if opt == "users" {
		// dump users, groups, group member
		if err := job.importTable(tx, zipReader, opt, nil); err != nil {
			return err
		}
		if err := job.importTable(tx, zipReader, "groups", nil); err != nil {
			return err
		}
		if err := job.importTable(tx, zipReader, "group_members", nil); err != nil {
			return err
		}
	} else if opt == "media" {
		// dump all local store files
		mediaHost := carrot.GetValue(job.m.db, models.KEY_CMS_MEDIA_HOST)
		mediaPrefix := carrot.GetValue(job.m.db, models.KEY_CMS_MEDIA_PREFIX)

		return job.importTable(tx, zipReader, opt, func(zr *zip.Reader, modelObj any) (bool, error) {
			media := modelObj.(*models.Media)
			if media.External || media.Directory {
				return true, nil
			}
			f, err := zr.Open(filepath.Join("media", media.StorePath))
			if err != nil {
				return false, err
			}
			defer f.Close()

			data := bytes.NewBuffer(nil)
			io.Copy(data, f)
			r, err := models.UploadFile(tx, media.Path, media.Name, data)
			if err != nil {
				return false, err
			}

			media.StorePath = r.StorePath
			media.Size = r.Size
			media.ContentType = r.ContentType
			media.Ext = r.Ext
			media.External = r.External
			media.Dimensions = r.Dimensions

			media.Directory = false
			media.Published = true
			media.CreatorID = job.user.ID

			media.BuildPublicUrls(mediaHost, mediaPrefix)
			media.Thumbnail = r.Thumbnail
			return true, nil
		})
	} else if opt == "pages" || opt == "posts" {
		origMediaHost := strings.TrimSuffix(job.Meta.MediaHost, "/")
		origMediaHost += job.Meta.MediaPrefix

		newMediaHost := strings.TrimSuffix(carrot.GetValue(job.m.db, models.KEY_CMS_MEDIA_HOST), "/")
		newMediaPrefix := carrot.GetValue(job.m.db, models.KEY_CMS_MEDIA_PREFIX)
		newMediaHost += newMediaPrefix

		return job.importTable(tx, zipReader, opt, func(zr *zip.Reader, modelObj any) (bool, error) {
			if origMediaHost == newMediaHost {
				return true, nil
			}
			if opt == "pages" {
				page := modelObj.(*models.Page)
				page.CreatorID = job.user.ID
				page.Thumbnail = strings.ReplaceAll(page.Thumbnail, origMediaHost, newMediaHost)
				page.Body = strings.ReplaceAll(page.Body, origMediaHost, newMediaHost)
				page.Draft = strings.ReplaceAll(page.Draft, origMediaHost, newMediaHost)
			} else {
				post := modelObj.(*models.Post)
				post.CreatorID = job.user.ID
				post.Thumbnail = strings.ReplaceAll(post.Thumbnail, origMediaHost, newMediaHost)
				post.Body = strings.ReplaceAll(post.Body, origMediaHost, newMediaHost)
				post.Draft = strings.ReplaceAll(post.Draft, origMediaHost, newMediaHost)
			}
			return true, nil
		})
	}
	return job.importTable(tx, zipReader, opt, nil)
}

func getAdminObject(db *gorm.DB, opt string) (*carrot.AdminObject, error) {
	obj := &carrot.AdminObject{
		Name: opt,
	}
	switch opt {
	case "users":
		obj.Model = &carrot.User{}
	case "groups":
		obj.Model = &carrot.Group{}
	case "group_members":
		obj.Model = &carrot.GroupMember{}
	case "sites":
		obj.Model = &models.Site{}
	case "categories":
		obj.Model = &models.Category{}
	case "pages":
		obj.Model = &models.Page{}
	case "posts":
		obj.Model = &models.Post{}
	case "media":
		obj.Model = &models.Media{}
	}

	err := obj.Build(db)
	return obj, err
}

func (job *ExportJob) dumpTable(out *zip.Writer, opt string, rowHandle rowExportHandle) (int, int64, error) {
	obj, err := getAdminObject(job.m.db, opt)
	if err != nil {
		return 0, 0, err
	}

	modelElem := reflect.TypeOf(obj.Model).Elem()
	vals := reflect.New(reflect.SliceOf(modelElem))
	result := vals.Interface()
	r := job.m.db.Model(obj.Model).Preload(clause.Associations).Find(result)
	if r.Error != nil {
		return 0, 0, r.Error
	}

	var lines []map[string]any
	var size int64
	for i := 0; i < vals.Elem().Len(); i++ {
		modelObj := vals.Elem().Index(i).Addr().Interface()
		item, err := obj.MarshalOne(modelObj)
		if err != nil {
			return 0, 0, err
		}
		if rowHandle != nil {
			rowSize, ok, err := rowHandle(out, modelObj)
			if !ok {
				continue
			}
			if err != nil {
				carrot.Warning("Dump row failed: ", modelObj, err)
				continue
			}
			size += rowSize
		}
		lines = append(lines, item)
	}

	name := fmt.Sprintf("%s.json", opt)
	f, err := out.Create(name)
	if err != nil {
		return 0, 0, err
	}
	data, err := json.Marshal(lines)
	if err != nil {
		return 0, 0, err
	}

	f.Write(data)
	return len(lines), size + int64(len(data)), nil
}

func (job *ExportJob) Dump(out *zip.Writer, opt string) (int, int64, error) {

	if opt == "users" {
		// dump users, groups, group member
		count, size, _ := job.dumpTable(out, opt, nil)
		groupCount, groupSize, _ := job.dumpTable(out, "groups", nil)
		memberCount, memberSize, _ := job.dumpTable(out, "group_members", nil)
		return count + groupCount + memberCount, size + groupSize + memberSize, nil
	} else if opt == "media" {
		// dump all local store files
		uploadDir := carrot.GetValue(job.m.db, models.KEY_CMS_UPLOAD_DIR)

		return job.dumpTable(out, opt, func(out *zip.Writer, modelObj any) (int64, bool, error) {
			media := modelObj.(*models.Media)
			if media.External || media.Directory {
				return 0, true, nil
			}

			if (strings.HasPrefix(media.Name, "restcontent_export_") || strings.HasPrefix(media.Name, "restcontent_backup_")) && strings.HasSuffix(media.Name, ".zip") {
				// ignore exported or backup file
				return 0, false, nil
			}

			fullPath := filepath.Join(uploadDir, media.StorePath)
			fileData, err := os.ReadFile(fullPath)
			if err != nil {
				return 0, false, err
			}
			f, err := out.Create(filepath.Join("media", media.StorePath))
			if err != nil {
				return 0, false, err
			}
			f.Write(fileData)
			return int64(len(fileData)), true, nil
		})
	}
	return job.dumpTable(out, opt, nil)
}

func (job *ExportJob) Start() {
	job.result.Status = "pending"

	go func() {
		defer func() {
			if err := recover(); err != nil {
				job.mutex.Lock()
				job.result.Status = "error"
				job.result.Reason = fmt.Sprintf("recover: %v", err)
				job.mutex.Unlock()
				carrot.Warning("Export job crash:", err)
			}
		}()
		defer time.AfterFunc(5*time.Second, func() {
			job.m.exportAndImportJobs.Delete(job.key)
		})

		exportMeta := ExportMeta{
			BuildTime:   job.m.BuildTime,
			Options:     []ExportOption{},
			From:        job.From,
			MediaHost:   job.MediaHost,
			MediaPrefix: carrot.GetValue(job.m.db, models.KEY_CMS_MEDIA_PREFIX),
			ExportTime:  time.Now(),
			Author:      carrot.GetValue(job.m.db, carrot.KEY_SITE_ADMIN),
		}

		zipFile := bytes.NewBuffer(nil)
		out := zip.NewWriter(zipFile)
		for _, opt := range job.Options {
			switch opt {
			case "users", "sites", "categories", "pages", "posts", "media":
			default:
				continue
			}
			count, size, err := job.Dump(out, opt)
			if err != nil {
				job.mutex.Lock()
				job.result.Status = "error"
				job.result.Reason = fmt.Sprintf("[%s] dump %v", opt, err.Error())
				job.mutex.Unlock()
				out.Close()
				return
			}

			metaOption := ExportOption{
				Name:  opt,
				Count: count,
				Size:  size,
			}
			exportMeta.Options = append(exportMeta.Options, metaOption)
		}

		metaData, _ := json.Marshal(&exportMeta)
		meta, _ := out.Create("meta.json")
		meta.Write([]byte(metaData))
		out.Close()
		// Save to media
		r, err := models.UploadFile(job.m.db, "/", fmt.Sprintf("restcontent_export_%s.zip", job.key), zipFile)
		if err != nil {
			job.mutex.Lock()
			job.result.Status = "error"
			job.result.Reason = fmt.Sprintf("UploadFile %v", err.Error())
			job.mutex.Unlock()
		} else {
			var media models.Media
			media.Name = r.Name
			media.Path = r.Path
			media.StorePath = r.StorePath
			media.Directory = false
			media.Ext = r.Ext
			media.ContentType = r.ContentType
			media.Published = true
			media.Size = r.Size

			media.CreatorID = job.user.ID
			media.Creator = *job.user

			result := job.m.db.Create(&media)
			if result.Error != nil {
				job.mutex.Lock()
				job.result.Status = "error"
				job.result.Reason = "create result: " + result.Error.Error()
				job.mutex.Unlock()
				return
			}

			mediaHost := carrot.GetValue(job.m.db, models.KEY_CMS_MEDIA_HOST)
			mediaPrefix := carrot.GetValue(job.m.db, models.KEY_CMS_MEDIA_PREFIX)
			media.BuildPublicUrls(mediaHost, mediaPrefix)

			job.mutex.Lock()
			job.result.Status = "done"
			job.result.DownloadLink = media.PublicUrl
			job.result.DownloadSize = media.Size
			job.mutex.Unlock()
		}
	}()
}

func (job *ExportJob) GetResult() ExportResult {
	job.mutex.Lock()
	defer job.mutex.Unlock()
	r := job.result
	return r
}

func (m *Manager) superAccessCheck(c *gin.Context) {
	if !carrot.CurrentUser(c).IsSuperUser {
		c.AbortWithError(403, errors.New("only superuser can access"))
		return
	}
	c.Next()
}

func (m *Manager) handleExportStart(c *gin.Context) {
	if !carrot.CurrentUser(c).IsSuperUser {
		carrot.AbortWithJSONError(c, 403, errors.New("permission denied"))
		return
	}

	var job ExportJob
	if err := c.ShouldBindJSON(&job); err != nil {
		carrot.AbortWithJSONError(c, 400, err)
		return
	}

	job.mutex = sync.Mutex{}
	job.m = m
	job.user = carrot.CurrentUser(c)
	n := time.Now()

	job.From = carrot.GetValue(m.db, carrot.KEY_SITE_URL)
	if job.From == "" {
		job.From = c.Request.Host
	}
	job.key = fmt.Sprintf("%s-%s-%s", job.From, n.Format("2006-01-02"), carrot.RandText(10))

	mediaHost := carrot.GetValue(m.db, models.KEY_CMS_MEDIA_HOST)
	if mediaHost == "" {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		mediaHost = scheme + "://" + c.Request.Host
	}
	job.MediaHost = mediaHost

	m.exportAndImportJobs.Store(job.key, &job)

	c.JSON(200, gin.H{"status": "pending", "key": job.key})
	job.Start()
}

func (m *Manager) handleExportPoll(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, errors.New("not key"))
		return
	}

	job, ok := m.exportAndImportJobs.Load(key)
	if !ok {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, errors.New("invalid job"))
		return
	}

	r := job.(*ExportJob).GetResult()
	if r.Status == "done" || r.Status == "error" {
		m.exportAndImportJobs.Delete(key)
	}

	c.JSON(200, r)
}

func (m *Manager) handleImportUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	key := "import_" + carrot.RandText(12)
	// write file to tmp file
	tmpFile, err := os.CreateTemp("", "restcontent_import_*.zip")
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	zipFile, err := file.Open()
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}
	defer zipFile.Close()

	// copy to tmp file
	if _, err := io.Copy(tmpFile, zipFile); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	zipFile.Seek(0, io.SeekStart) // reset to start
	zipReader, err := zip.NewReader(zipFile, file.Size)
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}
	meta, err := zipReader.Open("meta.json")
	if err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("open meta.json failed: %v", err))
		return
	}
	defer meta.Close()

	data := bytes.NewBuffer(nil)
	io.Copy(data, meta)

	var exportMeta ExportMeta
	if err := json.Unmarshal(data.Bytes(), &exportMeta); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, fmt.Errorf("parse meta.json failed: %v", err))
		return
	}
	exportMeta.Key = key
	job := ImportJob{
		m:           m,
		Meta:        exportMeta,
		key:         key,
		TmpFileName: tmpFile.Name(),
	}

	m.exportAndImportJobs.Store(key, &job)
	c.JSON(200, exportMeta)

	time.AfterFunc(1*time.Hour, func() {
		m.exportAndImportJobs.Delete(key)
	})
}

func (m *Manager) handleImportStart(c *gin.Context) {
	var form StartImportForm
	if err := c.ShouldBindJSON(&form); err != nil {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, err)
		return
	}

	obj, ok := m.exportAndImportJobs.Load(form.Key)
	if !ok {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, errors.New("invalid job"))
		return
	}

	job := obj.(*ImportJob)
	job.user = carrot.CurrentUser(c)

	r := job.GetResult()
	if r.Status == "done" || r.Status == "error" {
		m.exportAndImportJobs.Delete(form.Key)
	} else {
		job.Start(form.Options)
	}
	c.JSON(200, job.GetResult())
}

func (m *Manager) handleImportPoll(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, errors.New("not key"))
		return
	}

	job, ok := m.exportAndImportJobs.Load(key)
	if !ok {
		carrot.AbortWithJSONError(c, http.StatusBadRequest, errors.New("invalid job"))
		return
	}

	r := job.(*ImportJob).GetResult()
	if r.Status == "done" || r.Status == "error" {
		m.exportAndImportJobs.Delete(key)
	}
	c.JSON(200, r)
}
