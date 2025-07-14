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
	qwenCompatibleBatchesPath             = "/compatible-mode/v1/batches"
	qwenCompatibleRetrieveBatchPath       = "/compatible-mode/v1/batches/{batch_id}"
)

type qwenFile struct {
	provider.CommonRequestHandler
}

func (h *qwenFile) ShouldHandleRequest(ctx *gin.Context) bool {
	path := ctx.Request.URL.Path
	if strings.HasPrefix(path, qwenCompatibleFilesPath) || strings.HasPrefix(path, qwenCompatibleBatchesPath) {
		return true
	}

	return false
}

func (h *qwenFile) HandleFiles(c *gin.Context) {
}
