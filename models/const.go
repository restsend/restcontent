package models

import (
	"errors"

	"github.com/restsend/carrot"
)

const ENV_CMS_API_PREFIX = "CMS_API_PREFIX"
const KEY_CMS_GUEST_ACCESS_API = "CMS_GUEST_ACCESS_API"
const KEY_CMS_UPLOAD_DIR = "CMS_UPLOAD_DIR"
const KEY_CMS_EXTERNAL_UPLOADER = "CMS_EXTERNAL_UPLOADER"
const KEY_CMS_MEDIA_PREFIX = "CMS_MEDIA_PREFIX"
const KEY_CMS_MEDIA_HOST = "CMS_MEDIA_HOST"
const KEY_CMS_API_HOST = "CMS_API_HOST"
const KEY_CMS_RELATION_COUNT = "CMS_RELATION_COUNT"
const KEY_CMS_SUGGESTION_COUNT = "CMS_SUGGESTION_COUNT"

var ErrUnauthorized = errors.New("unauthorized")
var ErrDraftIsInvalid = errors.New("draft is invalid")
var ErrPageIsNotPublish = errors.New("page is not publish")
var ErrPostIsNotPublish = errors.New("post is not publish")
var ErrInvalidPathAndName = errors.New("invalid path and name")
var ErrUploadsDirNotConfigured = errors.New("uploads dir not configured")

const (
	ContentTypeHtml     = "html"
	ContentTypeJson     = "json"
	ContentTypeText     = "text"
	ContentTypeMarkdown = "markdown"
	ContentTypeImage    = "image"
	ContentTypeVideo    = "video"
	ContentTypeAudio    = "audio"
	ContentTypeFile     = "file"
)
const (
	DefaultCategoryUUIDSize = 12
	DefaultPageIDSize       = 14
)

var ContentTypes = []carrot.AdminSelectOption{
	{Value: ContentTypeJson, Label: "JSON"},
	{Value: ContentTypeHtml, Label: "HTML"},
	{Value: ContentTypeText, Label: "PlainText"},
	{Value: ContentTypeMarkdown, Label: "Markdown"},
	{Value: ContentTypeImage, Label: "Image"},
	{Value: ContentTypeVideo, Label: "Video"},
	{Value: ContentTypeAudio, Label: "Audio"},
	{Value: ContentTypeFile, Label: "File"},
}
