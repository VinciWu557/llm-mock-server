package embeddings

import (
	"net/http"

	"llm-mock-server/pkg/provider"

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
)

func HandleEmbeddings(context *gin.Context) {
	for _, handler := range embeddingsHandlers {
		if handler.ShouldHandleRequest(context) {
			handler.HandleEmbeddings(context)
			return
		}
	}
	context.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
}
