package batch

import (
	"llm-mock-server/pkg/provider"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	qwenCompatibleBatchesPath       = "/compatible-mode/v1/batches"
	qwenCompatibleRetrieveBatchPath = "/compatible-mode/v1/batches/{batch_id}"
)

var (
	qwenCompatibleRetrieveBatchPathRegex = regexp.MustCompile(`^/compatible-mode/v1/batches/(?P<batch_id>[^/]+)$`)
)

type qwenBatch struct {
	provider.CommonRequestHandler
}

func (h *qwenBatch) ShouldHandleRequest(ctx *gin.Context) bool {
	path := ctx.Request.URL.Path
	return strings.HasPrefix(path, qwenCompatibleBatchesPath)
}

func (h *qwenBatch) HandleBatches(c *gin.Context) {
	path := c.Request.URL.Path
	method := c.Request.Method

	if path == qwenCompatibleBatchesPath {
		h.handleBatches(c, method)
		return
	}

	if matches := qwenCompatibleRetrieveBatchPathRegex.FindStringSubmatch(path); len(matches) > 0 {
		batchID := getNamedCaptureValue(qwenCompatibleRetrieveBatchPathRegex, matches, "batch_id")
		h.handleSingleBatch(c, method, batchID)
		return
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "path not found"})
}

func (h *qwenBatch) handleBatches(c *gin.Context, method string) {
	switch method {
	case http.MethodPost:
		var req createBatchRequest
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, createBatchResponse(req.InputFileId))
	case http.MethodGet:
		c.JSON(http.StatusOK, createQwenBatchListResponse())
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}

func (h *qwenBatch) handleSingleBatch(c *gin.Context, method string, batchID string) {
	switch method {
	case http.MethodGet:
		c.JSON(http.StatusOK, createBatchResponse(batchID))
	default:
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "method not allowed"})
	}
}
