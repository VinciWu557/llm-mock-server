package batch

import (
	"llm-mock-server/pkg/provider"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	batchMockID          string = "batch-abc123"
	batchMockCreated     int64  = 10
	batchMockInputFileId string = "file-abc123"
	batchMockCustomerID  string = "user_123456789"
	batchMockDescription string = "Nightly eval job"

	openaiBatchesPath       = "/v1/batches"
	openaiRetrieveBatchPath = "/v1/batches/{batch_id}"
)

var (
	RegRetrieveBatchPath = regexp.MustCompile(`^/v1/batches/(?P<batch_id>[^/]+)$`)
)

type openAIBatch struct {
	provider.CommonRequestHandler
}

func (h *openAIBatch) ShouldHandleRequest(ctx *gin.Context) bool {
	path := ctx.Request.URL.Path
	return strings.HasPrefix(path, "/v1/batches")
}

func (h *openAIBatch) HandleBatches(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	// openaiBatchesPath
	if path == openaiBatchesPath {
		h.handleBatches(c, method)
		return
	}

	// openaiRetrieveBatchPath
	if matches := RegRetrieveBatchPath.FindStringSubmatch(path); len(matches) > 0 {
		batchID := getNamedCaptureValue(RegRetrieveBatchPath, matches, "batch_id")
		h.handleSingleBatch(c, method, batchID)
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

func (h *openAIBatch) handleBatches(c *gin.Context, method string) {
	switch method {
	case http.MethodPost:
		var req createBatchRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, createBatchResponse(req.InputFileId))
	case http.MethodGet:
		c.JSON(http.StatusOK, createBatchListResponse())
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

func (h *openAIBatch) handleSingleBatch(c *gin.Context, method string, batchID string) {
	switch method {
	case http.MethodGet:
		c.JSON(http.StatusOK, createBatchResponse(batchID))
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

// createBatchResponse 创建标准的 batch 响应结构体
func createBatchResponse(batchID string) batch {
	return batch{
		Id:       batchID,
		Object:   "batch",
		Endpoint: "/v1/chat/completions",
		Errors: struct {
			Data []struct {
				Code    string `json:"code"`
				Line    int    `json:"line"`
				Message string `json:"message"`
				Param   string `json:"param"`
			} `json:"data"`
			Object string `json:"object"`
		}{},
		InputFileId:      batchMockInputFileId,
		CompletionWindow: "24h",
		Status:           "validating",
		OutputFileId:     "",
		ErrorFileId:      "",
		CreatedAt:        batchMockCreated,
		InProgressAt:     0,
		ExpiresAt:        0,
		FinalizingAt:     0,
		CompletedAt:      0,
		FailedAt:         0,
		ExpiredAt:        0,
		CancellingAt:     0,
		CancelledAt:      0,
		RequestCounts: struct {
			Completed int `json:"completed"`
			Failed    int `json:"failed"`
			Total     int `json:"total"`
		}{
			Completed: 0,
			Failed:    0,
			Total:     0,
		},
		Metadata: map[string]string{
			"customer_id":       batchMockCustomerID,
			"batch_description": batchMockDescription,
		},
	}
}

// createBatchListResponse 创建标准的 batch 列表响应
func createBatchListResponse() gin.H {
	return gin.H{
		"object": "list",
		"data": []*batch{
			{
				Id:       batchMockID,
				Object:   "batch",
				Endpoint: "/v1/chat/completions",
				Errors: struct {
					Data []struct {
						Code    string `json:"code"`
						Line    int    `json:"line"`
						Message string `json:"message"`
						Param   string `json:"param"`
					} `json:"data"`
					Object string `json:"object"`
				}{},
				InputFileId:      batchMockInputFileId,
				CompletionWindow: "24h",
				Status:           "validating",
				OutputFileId:     "",
				ErrorFileId:      "",
				CreatedAt:        batchMockCreated,
				InProgressAt:     0,
				ExpiresAt:        0,
				FinalizingAt:     0,
				CompletedAt:      0,
				FailedAt:         0,
				ExpiredAt:        0,
				CancellingAt:     0,
				CancelledAt:      0,
				RequestCounts: struct {
					Completed int `json:"completed"`
					Failed    int `json:"failed"`
					Total     int `json:"total"`
				}{
					Completed: 0,
					Failed:    0,
					Total:     0,
				},
				Metadata: map[string]string{
					"customer_id":       batchMockCustomerID,
					"batch_description": batchMockDescription,
				},
			},
		},
		"first_id": batchMockID,
		"last_id":  batchMockID,
		"has_more": true,
	}
}

// createQwenBatchListResponse 创建 qwen 兼容的简化版 batch 列表响应
func createQwenBatchListResponse() gin.H {
	return gin.H{
		"object": "list",
		"data": []*batch{
			{
				Id:               batchMockID,
				Object:           "batch",
				Endpoint:         "https://api.qwen.com/v1/batches",
				CompletionWindow: "1000",
				Metadata: map[string]string{
					"key": "value",
				},
			},
		},
	}
}

type batch struct {
	Id       string `json:"id"`
	Object   string `json:"object"`
	Endpoint string `json:"endpoint"`
	Errors   struct {
		Data []struct {
			Code    string `json:"code"`
			Line    int    `json:"line"`
			Message string `json:"message"`
			Param   string `json:"param"`
		} `json:"data"`
		Object string `json:"object"`
	} `json:"errors"`
	InputFileId      string `json:"input_file_id"`
	CompletionWindow string `json:"completion_window"`
	Status           string `json:"status"`
	OutputFileId     string `json:"output_file_id"`
	ErrorFileId      string `json:"error_file_id"`
	CreatedAt        int64  `json:"created_at"`
	InProgressAt     int64  `json:"in_progress_at"`
	ExpiresAt        int64  `json:"expires_at"`
	FinalizingAt     int64  `json:"finalizing_at"`
	CompletedAt      int64  `json:"completed_at"`
	FailedAt         int64  `json:"failed_at"`
	ExpiredAt        int64  `json:"expired_at"`
	CancellingAt     int64  `json:"cancelling_at"`
	CancelledAt      int64  `json:"cancelled_at"`
	RequestCounts    struct {
		Completed int `json:"completed"`
		Failed    int `json:"failed"`
		Total     int `json:"total"`
	} `json:"request_counts"`
	Metadata map[string]string `json:"metadata"`
}

type createBatchRequest struct {
	InputFileId      string            `json:"input_file_id" binding:"required"`
	Endpoint         string            `json:"endpoint" binding:"required"`
	CompletionWindow string            `json:"completion_window" binding:"required"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}
