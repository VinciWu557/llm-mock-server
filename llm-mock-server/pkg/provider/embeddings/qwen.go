package embeddings

import (
	"llm-mock-server/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	qwenDomain                      = "dashscope.aliyuncs.com"
	qwenTextEmbeddingPath           = "/api/v1/services/embeddings/text-embedding/text-embedding"
	qwenCompatibleTextEmbeddingPath = "/compatible-mode/v1/embeddings"
)

type qwenEmbeddings struct {
}

func (h *qwenEmbeddings) ShouldHandleRequest(ctx *gin.Context) bool {
	context, _ := utils.GetRequestContext(ctx)

	if context.Host != qwenDomain {
		return false
	}

	supportedPaths := []string{qwenTextEmbeddingPath, qwenCompatibleTextEmbeddingPath}
	for _, path := range supportedPaths {
		if context.Path == path {
			return true
		}
	}

	return false
}

func (h *qwenEmbeddings) HandleEmbeddings(c *gin.Context) {
}

type qwenTextEmbeddingRequest struct {
	Model      string                      `json:"model"`
	Input      qwenTextEmbeddingInput      `json:"input"`
	Parameters qwenTextEmbeddingParameters `json:"parameters,omitempty"`
}

type qwenTextEmbeddingInput struct {
	Texts []string `json:"texts"`
}

type qwenTextEmbeddingParameters struct {
	TextType string `json:"text_type,omitempty"`
}

type qwenTextEmbeddingResponse struct {
	RequestId string                  `json:"request_id"`
	Output    qwenTextEmbeddingOutput `json:"output"`
	Usage     qwenUsage               `json:"usage"`
}

type qwenTextEmbeddingOutput struct {
	RequestId  string               `json:"request_id"`
	Embeddings []qwenTextEmbeddings `json:"embeddings"`
}

type qwenTextEmbeddings struct {
	TextIndex int       `json:"text_index"`
	Embedding []float64 `json:"embedding"`
}

type qwenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
