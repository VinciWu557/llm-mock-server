package file

import (
	"llm-mock-server/pkg/provider"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	fileMockCreated       int64  = 10
	fileMockID            string = "file-abc123"
	fileMockFilename      string = "test.txt"
	fileMockPurpose       string = "assistants"
	fileMockStatus        string = "processed"
	fileMockBytes         int    = 140
	fileMockStatusDetails string = "test_status_details"

	openaiFilesPath               = "/v1/files"
	openaiRetrieveFilePath        = "/v1/files/{file_id}"
	openaiRetrieveFileContentPath = "/v1/files/{file_id}/content"
)

var (
	RegRetrieveFilePath        = regexp.MustCompile(`^/v1/files/(?P<file_id>[^/]+)$`)
	RegRetrieveFileContentPath = regexp.MustCompile(`^/v1/files/(?P<file_id>[^/]+)/content$`)
)

type openaiFile struct {
	provider.CommonRequestHandler
}

func (h *openaiFile) ShouldHandleRequest(ctx *gin.Context) bool {
	path := ctx.Request.URL.Path
	return strings.HasPrefix(path, "/v1/files")
}

func (h *openaiFile) HandleFiles(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	// openaiFilesPath
	if path == openaiFilesPath {
		h.handleFiles(c, method)
		return
	}

	// openaiRetrieveFilePath
	if matches := RegRetrieveFilePath.FindStringSubmatch(path); len(matches) > 0 {
		fileID := getNamedCaptureValue(RegRetrieveFilePath, matches, "file_id")
		h.handleSingleFile(c, method, fileID)
		return
	}

	// openaiRetrieveFileContentPath
	if matches := RegRetrieveFileContentPath.FindStringSubmatch(path); len(matches) > 0 {
		fileID := getNamedCaptureValue(RegRetrieveFileContentPath, matches, "file_id")
		h.handleFileContent(c, method, fileID)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "path not found"})
}

func getNamedCaptureValue(r *regexp.Regexp, matches []string, name string) string {
	index := r.SubexpIndex(name)
	if index >= 0 && index < len(matches) {
		return matches[index]
	}
	return ""
}

func (h *openaiFile) handleFiles(c *gin.Context, method string) {
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

func (h *openaiFile) handleSingleFile(c *gin.Context, method string, fileID string) {
	switch method {
	case http.MethodGet:
		c.JSON(http.StatusOK, createFileResponse(fileID))
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

func (h *openaiFile) handleFileContent(c *gin.Context, method string, fileID string) {
	switch method {
	case http.MethodGet:
		c.String(http.StatusOK, createFileContentResponse(fileID))
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

// createUploadFileResponse 创建文件上传响应
func createUploadFileResponse() uploadFileResponse {
	return uploadFileResponse{
		Id:        fileMockID,
		Object:    "file",
		Bytes:     fileMockBytes,
		CreatedAt: fileMockCreated,
		Filename:  fileMockFilename,
		Purpose:   fileMockPurpose,
	}
}

// createFileListResponse 创建文件列表响应
func createFileListResponse() gin.H {
	return gin.H{
		"object": "list",
		"data": []*file{
			{
				Id:            fileMockID,
				Object:        "file",
				Bytes:         fileMockBytes,
				CreatedAt:     fileMockCreated,
				Filename:      fileMockFilename,
				Purpose:       fileMockPurpose,
				Status:        fileMockStatus,
				StatusDetails: fileMockStatusDetails,
			},
		},
	}
}

// createFileResponse 创建单个文件响应
func createFileResponse(fileID string) file {
	return file{
		Id:            fileID,
		Object:        "file",
		Bytes:         fileMockBytes,
		CreatedAt:     fileMockCreated,
		Filename:      fileMockFilename,
		Purpose:       fileMockPurpose,
		Status:        fileMockStatus,
		StatusDetails: fileMockStatusDetails,
	}
}

// createFileContentResponse 创建文件内容响应
func createFileContentResponse(fileID string) string {
	return fileID
}

type file struct {
	Id            string `json:"id"`
	Object        string `json:"object"`
	Bytes         int    `json:"bytes"`
	CreatedAt     int64  `json:"created_at"`
	ExpiresAt     int64  `json:"expires_at"`
	Filename      string `json:"filename"`
	Purpose       string `json:"purpose"`
	Status        string `json:"status"`
	StatusDetails string `json:"status_details"`
}

type uploadFileRequest struct {
	Purpose string `form:"purpose"`
}

type uploadFileResponse struct {
	Id        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}
