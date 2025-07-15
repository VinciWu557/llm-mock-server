package file

import (
	"llm-mock-server/pkg/provider"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	qwenCompatibleFilesPath               = "/compatible-mode/v1/files"
	qwenCompatibleRetrieveFilePath        = "/compatible-mode/v1/files/{file_id}"
	qwenCompatibleRetrieveFileContentPath = "/compatible-mode/v1/files/{file_id}/content"
)

type qwenFile struct {
	provider.CommonRequestHandler
}

func (h *qwenFile) ShouldHandleRequest(ctx *gin.Context) bool {
	path := ctx.Request.URL.Path
	return strings.HasPrefix(path, qwenCompatibleFilesPath)
}

func (h *qwenFile) HandleFiles(c *gin.Context) {
}
