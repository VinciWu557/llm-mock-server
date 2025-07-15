package batch

import (
	"llm-mock-server/pkg/provider"
	"llm-mock-server/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BatchRequestHandler 定义批处理请求处理器接口
type BatchRequestHandler interface {
	provider.CommonRequestHandler
	HandleBatches(ctx *gin.Context)
}

var (
	batchHandlers = []BatchRequestHandler{
		&qwenBatch{},
		&openAIBatch{}, // 作为最后的回退处理器
	}

	batchRoutes = []string{
		// qwen
		qwenCompatibleBatchesPath,
		qwenCompatibleRetrieveBatchPath,
		// openai
		openaiBatchesPath,
		openaiRetrieveBatchPath,
	}
)

// SetupRoutes 设置批处理相关的路由
func SetupRoutes(server *gin.Engine) {
	for _, route := range batchRoutes {
		server.Any(route, handleBatch)
	}
}

func handleBatch(context *gin.Context) {
	if err := utils.BuildRequestContext(context); err != nil {
		return
	}

	for _, handler := range batchHandlers {
		if handler.ShouldHandleRequest(context) {
			handler.HandleBatches(context)
			return
		}
	}
	context.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
}

// GetBatchRequestHandler 返回批处理请求处理器
func GetBatchRequestHandler() BatchRequestHandler {
	return &openAIBatch{}
}
