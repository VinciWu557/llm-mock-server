package embeddings

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
	embeddingsHandlers = []requestHandler{
		&qwenEmbeddings{},
		&openaiEmbeddings{},
	}

	embeddingsRoutes = []string{
		// qwen
		"/compatible-mode/v1/embeddings",
		"/api/v1/services/embeddings/text-embedding/text-embedding",
		// openai
		"/v1/embeddings",
	}
)

func SetupRoutes(server *gin.Engine) {
	for _, route := range embeddingsRoutes {
		server.POST(route, handleEmbeddings)
	}
}

func handleEmbeddings(context *gin.Context) {
	if err := utils.BuildRequestContext(context); err != nil {
		return
	}

	for _, handler := range embeddingsHandlers {
		if handler.ShouldHandleRequest(context) {
			handler.HandleEmbeddings(context)
			return
		}
	}
	context.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
}
