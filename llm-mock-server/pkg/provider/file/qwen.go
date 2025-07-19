package file

import (
	"llm-mock-server/pkg/provider"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	qwenCompatibleFilesPath               = "/compatible-mode/v1/files"
	qwenCompatibleRetrieveFilePath        = "/compatible-mode/v1/files/{file_id}"
	qwenCompatibleRetrieveFileContentPath = "/compatible-mode/v1/files/{file_id}/content"
)

var (
	qwenCompatibleRetrieveFilePathRegex        = regexp.MustCompile(`^/compatible-mode/v1/files/(?P<file_id>[^/]+)$`)
	qwenCompatibleRetrieveFileContentPathRegex = regexp.MustCompile(`^/compatible-mode/v1/files/(?P<file_id>[^/]+)/content$`)
)

type qwenFile struct {
	provider.CommonRequestHandler
}

func (h *qwenFile) ShouldHandleRequest(ctx *gin.Context) bool {
	path := ctx.Request.URL.Path
	return strings.HasPrefix(path, qwenCompatibleFilesPath)
}

func (h *qwenFile) HandleFiles(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	if path == qwenCompatibleFilesPath {
		h.handleFiles(c, method)
		return
	}

	if matches := qwenCompatibleRetrieveFilePathRegex.FindStringSubmatch(path); len(matches) > 0 {
		fileID := getNamedCaptureValue(qwenCompatibleRetrieveFilePathRegex, matches, "file_id")
		h.handleSingleFile(c, method, fileID)
		return
	}

	if matches := qwenCompatibleRetrieveFileContentPathRegex.FindStringSubmatch(path); len(matches) > 0 {
		fileID := getNamedCaptureValue(qwenCompatibleRetrieveFileContentPathRegex, matches, "file_id")
		h.handleFileContent(c, method, fileID)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "path not found"})
}

func (h *qwenFile) handleFiles(c *gin.Context, method string) {
	switch method {
	case http.MethodPost:
		var req uploadFileRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, createUploadFileResponse())
	case http.MethodGet:
		c.JSON(http.StatusOK, createFileListResponse())
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

func (h *qwenFile) handleSingleFile(c *gin.Context, method string, fileID string) {
	switch method {
	case http.MethodGet:
		c.JSON(http.StatusOK, createFileResponse(fileID))
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

func (h *qwenFile) handleFileContent(c *gin.Context, method string, fileID string) {
	switch method {
	case http.MethodGet:
		c.String(http.StatusOK, createFileContentResponse(fileID))
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}
