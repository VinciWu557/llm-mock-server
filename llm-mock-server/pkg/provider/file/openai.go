package file

import (
	"llm-mock-server/pkg/provider"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	openaiFilesPath               = "/v1/files"
	openaiRetrieveFilePath        = "/v1/files/{file_id}"
	openaiRetrieveFileContentPath = "/v1/files/{file_id}/content"
	openaiBatchesPath             = "/v1/batches"
	openaiRetrieveBatchPath       = "/v1/batches/{batch_id}"
)

type openaiFile struct {
	provider.CommonRequestHandler
}

func (h *openaiFile) ShouldHandleRequest(ctx *gin.Context) bool {
	return strings.HasPrefix(ctx.Request.URL.Path, "/v1/files")
}

func (h *openaiFile) HandleFiles(c *gin.Context) {
}
