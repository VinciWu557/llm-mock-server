package file

import (
	"llm-mock-server/pkg/provider"
	"llm-mock-server/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type requestHandler interface {
	provider.CommonRequestHandler

	HandleFiles(context *gin.Context)
}

var (
	fileHandlers = []requestHandler{
		&qwenFile{},
		&openaiFile{},
	}

	fileRoutes = []string{
		// qwen
		qwenCompatibleFilesPath,
		qwenCompatibleRetrieveFilePath,
		qwenCompatibleRetrieveFileContentPath,
		qwenCompatibleBatchesPath,
		qwenCompatibleRetrieveBatchPath,
		// openai
		openaiFilesPath,
		openaiRetrieveFilePath,
		openaiRetrieveFileContentPath,
		openaiBatchesPath,
		openaiRetrieveBatchPath,
	}
)

func SetupRoutes(server *gin.Engine) {
	for _, route := range fileRoutes {
		server.POST(route, handleFile)
	}
}

func handleFile(context *gin.Context) {
	if err := utils.BuildRequestContext(context); err != nil {
		return
	}

	for _, handler := range fileHandlers {
		if handler.ShouldHandleRequest(context) {
			handler.HandleFiles(context)
			return
		}
	}
	context.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
}
