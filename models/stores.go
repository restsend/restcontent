package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/restsend/carrot"
	"gorm.io/gorm"
)

type UploadResult struct {
	PublicUrl   string `json:"publicUrl"`
	Thumbnail   string `json:"thumbnail"`
	Path        string `json:"path"`
	Name        string `json:"name"`
	External    bool   `json:"external"`
	StorePath   string `json:"storePath"`
	Dimensions  string `json:"dimensions"`
	Ext         string `json:"ext"`
	Size        int64  `json:"size"`
	ContentType string `json:"contentType"`
}

func RemoveDirectory(db *gorm.DB, path string) (string, error) {
	var files []Media
	r := db.Model(&Media{}).Where("path", path).Find(&files)
	if r.Error != nil {
		carrot.Warning("Remove directory failed: ", r.Error, path)
		return "", r.Error
	}

	uploadDir := carrot.GetValue(db, KEY_CMS_UPLOAD_DIR)
	for _, media := range files {
		if media.Directory {
			RemoveDirectory(db, filepath.Join(path, media.Name))
			continue
		}
		if !media.External {
			fullPath := filepath.Join(uploadDir, media.StorePath)
			if err := os.Remove(fullPath); err != nil {
				carrot.Warning("Remove file failed: ", err, fullPath)
			}
		}
	}

	r = db.Where("path", path).Delete(&Media{})
	if r.Error != nil {
		return "", r.Error
	}

	parent, name := filepath.Split(path)
	if parent != "/" {
		parent = strings.TrimSuffix(parent, "/")
	}
	return parent, db.Where("path", parent).Where("name", name).Delete(&Media{}).Error
}

func RemoveFile(db *gorm.DB, path, name string) error {
	if name == "" {
		return ErrInvalidPathAndName
	}

	media, err := GetMedia(db, path, name)
	if err != nil {
		return err
	}

	if !media.External {
		return nil
	}

	uploadDir := carrot.GetValue(db, KEY_CMS_UPLOAD_DIR)
	fullPath := filepath.Join(uploadDir, media.StorePath)
	if err := os.Remove(fullPath); err != nil {
		return err
	}
	return nil
}

func PrepareStoreLocalDir(db *gorm.DB) (string, error) {
	uploadDir := carrot.GetValue(db, KEY_CMS_UPLOAD_DIR)
	if uploadDir == "" {
		return "", ErrUploadsDirNotConfigured
	}

	if _, err := os.Stat(uploadDir); err != nil {
		if os.IsNotExist(err) {
			carrot.Warning("upload dir not exist, create it: ", uploadDir)
			if err = os.MkdirAll(uploadDir, 0755); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return uploadDir, nil
}

func StoreLocal(uploadDir, storePath string, data []byte) (string, error) {
	storePath = filepath.Join(uploadDir, storePath)
	err := os.WriteFile(storePath, data, 0644)
	if err != nil {
		return "", err
	}
	return storePath, nil
}

func StoreExternal(externalUploader, path, name string, data []byte) (string, error) {
	buf := new(bytes.Buffer)
	form := multipart.NewWriter(buf)
	form.WriteField("path", path)
	form.WriteField("name", name)

	fileField, _ := form.CreateFormFile("file", name)
	fileField.Write(data)
	form.Close()

	resp, err := http.Post(externalUploader, form.FormDataContentType(), buf)
	if err != nil {
		carrot.Warning("upload to external server failed: ", err, externalUploader)
		return "", err
	}

	defer resp.Body.Close()
	respBody := bytes.NewBuffer(nil)
	io.Copy(respBody, resp.Body)
	body := respBody.Bytes()
	if resp.StatusCode != http.StatusOK {
		carrot.Warning("upload to external server failed: ", resp.StatusCode, externalUploader, string(body))
		return "", fmt.Errorf("upload to external server failed, code:%d %s", resp.StatusCode, string(body))
	}
	var remoteResult UploadResult
	json.Unmarshal(body, &remoteResult)
	return remoteResult.StorePath, nil
}

func UploadFile(db *gorm.DB, path, name string, reader io.Reader) (*UploadResult, error) {
	var r UploadResult
	r.Path = path
	r.Name = name
	r.Ext = strings.ToLower(filepath.Ext(name))

	canGetDimension := false

	switch r.Ext {
	case ".jpg", ".jpeg", ".png", ".gif":
		canGetDimension = true
		fallthrough
	case ".webp", ".svg", ".ico", ".bmp":
		r.ContentType = ContentTypeImage
	case ".mp3", ".wav", ".ogg", ".aac", ".flac":
		r.ContentType = ContentTypeAudio
	case ".mp4", ".webm", ".avi", ".mov", ".wmv", ".mkv":
		r.ContentType = ContentTypeVideo
	default:
		r.ContentType = ContentTypeFile
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	r.Size = int64(len(data))

	externalUploader := carrot.GetValue(db, KEY_CMS_EXTERNAL_UPLOADER)
	if externalUploader != "" {
		storePath, err := StoreExternal(externalUploader, path, name, data)
		if err != nil {
			return nil, err
		}
		r.StorePath = storePath
		r.External = true
	} else {
		storePath := fmt.Sprintf("%s%s", carrot.RandText(10), r.Ext)
		r.StorePath = storePath
		r.External = false
		uploadDir, err := PrepareStoreLocalDir(db)
		if err != nil {
			return nil, err
		}
		_, err = StoreLocal(uploadDir, storePath, data)
		if err != nil {
			return nil, err
		}
	}

	if canGetDimension {
		config, _, err := image.DecodeConfig(bytes.NewReader(data))
		if err == nil {
			r.Dimensions = fmt.Sprintf("%dX%d", config.Width, config.Height)
		} else {
			carrot.Warning("decode image config error: ", err)
			r.Dimensions = "X"
		}
	}
	return &r, nil
}

func GetMedia(db *gorm.DB, path, name string) (*Media, error) {
	var obj Media
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	tx := db.Model(&Media{}).Where("path", path).Where("name", name)
	r := tx.First(&obj)
	if r.Error != nil {
		return nil, r.Error
	}
	return &obj, nil
}
