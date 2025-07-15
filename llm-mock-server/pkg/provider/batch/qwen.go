package batch

import (
	"llm-mock-server/pkg/provider"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	qwenCompatibleBatchesPath       = "/compatible-mode/v1/batches"
	qwenCompatibleRetrieveBatchPath = "/compatible-mode/v1/batches/{batch_id}"
)

type qwenBatch struct {
	provider.CommonRequestHandler
}

func (h *qwenBatch) ShouldHandleRequest(ctx *gin.Context) bool {
	path := ctx.Request.URL.Path
	return strings.HasPrefix(path, qwenCompatibleBatchesPath)
}

func (h *qwenBatch) HandleBatches(c *gin.Context) {
	// TODO 目前只是转发到 OpenAI 的处理逻辑
	(&openAIBatch{}).HandleBatches(c)
}
