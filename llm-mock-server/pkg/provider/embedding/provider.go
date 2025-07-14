package embedding

import (
	"net/http"

	"llm-mock-server/pkg/provider"
	"llm-mock-server/pkg/utils"

	"github.com/gin-gonic/gin"
)

type requestHandler interface {
	provider.CommonRequestHandler

	HandleEmbeddings(context *gin.Context)
}

var (
	embeddingHandlers = []requestHandler{
		&qwenEmbedding{},
		&openaiEmbedding{},
	}

	embeddingRoutes = []string{
		// qwen
		qwenTextEmbeddingPath,
		qwenCompatibleTextEmbeddingPath,
		// openai
		openaiEmbeddingPath,
	}
)

func SetupRoutes(server *gin.Engine) {
	for _, route := range embeddingRoutes {
		server.POST(route, handleEmbedding)
	}
}

func handleEmbedding(context *gin.Context) {
	if err := utils.BuildRequestContext(context); err != nil {
		return
	}

	for _, handler := range embeddingHandlers {
		if handler.ShouldHandleRequest(context) {
			handler.HandleEmbeddings(context)
			return
		}
	}
	context.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
}
