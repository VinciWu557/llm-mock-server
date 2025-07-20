package file

import (
	"llm-mock-server/pkg/provider"
	"mime/multipart"
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
	openaiRetrieveFilePath        = "/v1/files/:file_id"
	openaiRetrieveFileContentPath = "/v1/files/:file_id/content"
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

		c.JSON(http.StatusOK, createUploadFileResponse(req))
	case http.MethodGet:
		c.JSON(http.StatusOK, getFileListResponse())
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

func (h *openaiFile) handleSingleFile(c *gin.Context, method string, fileID string) {
	switch method {
	case http.MethodGet:
		c.JSON(http.StatusOK, retrieveFileResponse(fileID))
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

func (h *openaiFile) handleFileContent(c *gin.Context, method string, fileID string) {
	switch method {
	case http.MethodGet:
		c.JSON(http.StatusOK, retrieveFileContentResponse(fileID))
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

// createUploadFileResponse 创建文件上传响应
func createUploadFileResponse(req uploadFileRequest) uploadFileResponse {
	return uploadFileResponse{
		Id:        fileMockID,
		Object:    "file",
		Bytes:     fileMockBytes,
		CreatedAt: fileMockCreated,
		Filename:  req.File.Filename,
		Purpose:   req.Purpose,
	}
}

// getFileListResponse 获取文件列表响应
func getFileListResponse() gin.H {
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

// createFileResponse 检索单个文件响应
func retrieveFileResponse(fileID string) file {
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

// retrieveFileContentResponse 检索文件内容响应
func retrieveFileContentResponse(fileID string) fileContent {
	requestID := fileID + "-7616-97bf-87f2-7d747bbe84fd"

	return fileContent{
		Id:       fileID,
		CustomId: "1",
		Response: responseData{
			StatusCode: 200,
			RequestId:  requestID,
			Body: responseBody{
				Id:      "chatcmpl-" + requestID,
				Created: 1742303743,
				Usage: struct {
					CompletionTokens int `json:"completion_tokens"`
					PromptTokens     int `json:"prompt_tokens"`
					TotalTokens      int `json:"total_tokens"`
				}{
					CompletionTokens: 7,
					PromptTokens:     26,
					TotalTokens:      33,
				},
				Model: "qwen-max",
				Choices: []struct {
					FinishReason string `json:"finish_reason"`
					Index        int    `json:"index"`
					Message      struct {
						Content string `json:"content"`
					} `json:"message"`
				}{
					{
						FinishReason: "stop",
						Index:        0,
						Message: struct {
							Content string `json:"content"`
						}{
							Content: "2+2 equals 4.",
						},
					},
				},
				Object: "chat.completion",
			},
		},
		Error:            nil,
		CompletionTokens: 7,
		PromptTokens:     26,
		ReasoningTokens:  0,
		Model:            "qwen-max",
		ReasoningContent: "2+2 equals 4.",
	}
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
	File    *multipart.FileHeader `form:"file" binding:"required"`
	Purpose string                `form:"purpose" binding:"required"`
}

type uploadFileResponse struct {
	Id        string `json:"id"`
	Object    string `json:"object"`
	Bytes     int    `json:"bytes"`
	CreatedAt int64  `json:"created_at"`
	Filename  string `json:"filename"`
	Purpose   string `json:"purpose"`
}

// responseBody 响应体结构
type responseBody struct {
	Id      string `json:"id"`
	Created int64  `json:"created"`
	Usage   struct {
		CompletionTokens int `json:"completion_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Model   string `json:"model"`
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
		Message      struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Object string `json:"object"`
}

// responseData 响应数据结构
type responseData struct {
	StatusCode int          `json:"status_code"`
	RequestId  string       `json:"request_id"`
	Body       responseBody `json:"body"`
}

// errorData 错误数据结构
type errorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type fileContent struct {
	Id               string       `json:"id"`
	CustomId         string       `json:"custom_id"`
	Response         responseData `json:"response"`
	Error            *errorData   `json:"error"`
	CompletionTokens int          `json:"completion_tokens"`
	PromptTokens     int          `json:"prompt_tokens"`
	ReasoningTokens  int          `json:"reasoning_tokens"`
	Model            string       `json:"model"`
	ReasoningContent string       `json:"reasoning_content"`
}
