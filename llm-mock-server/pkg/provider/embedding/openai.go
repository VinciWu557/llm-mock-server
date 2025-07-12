package embedding

import (
	"fmt"
	"net/http"
	"reflect"

	"llm-mock-server/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var (
	embeddingMockUsage = usage{
		PromptTokens: 8,
		TotalTokens:  8,
	}

	// 简单的固定 mock embedding 向量，方便测试
	mockEmbeddingVector = []float64{0.1, 0.2, 0.3}
)

type openaiEmbedding struct {
}

func (e *openaiEmbedding) ShouldHandleRequest(ctx *gin.Context) bool {
	return ctx.Request.URL.Path == "/v1/embeddings"
}

func (e *openaiEmbedding) HandleEmbeddings(c *gin.Context) {
	// 验证 Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No API key provided"})
		return
	}

	// 绑定请求体
	var request embeddingsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request body: %v", err.Error())})
		return
	}

	// 验证请求体
	if err := utils.Validate.Struct(request); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		for _, fieldError := range validationErrors {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid parameter '%s': %v", fieldError.Field(), fieldError.Tag())})
			return
		}
	}

	// 处理输入文本
	texts, err := e.extractTextsFromInput(request.Input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 生成 mock embeddings 响应
	response := e.createEmbeddingsResponse(request, texts)
	c.JSON(http.StatusOK, response)
}

// 从输入中提取文本数组
func (e *openaiEmbedding) extractTextsFromInput(input interface{}) ([]string, error) {
	switch v := input.(type) {
	case string:
		return []string{v}, nil
	case []interface{}:
		texts := make([]string, 0, len(v))
		for i, item := range v {
			if str, ok := item.(string); ok {
				texts = append(texts, str)
			} else {
				return nil, fmt.Errorf("invalid input type at index %d: expected string, got %s",
					i, reflect.TypeOf(item).String())
			}
		}
		return texts, nil
	default:
		return nil, fmt.Errorf("invalid input type: expected string or array of strings, got %s",
			reflect.TypeOf(input).String())
	}
}

// 创建 embeddings 响应
func (e *openaiEmbedding) createEmbeddingsResponse(request embeddingsRequest, texts []string) embeddingsResponse {
	data := make([]embedding, len(texts))

	// 直接使用固定的 mock 向量
	for i := range texts {
		data[i] = embedding{
			Object:    "embedding",
			Index:     i,
			Embedding: mockEmbeddingVector,
		}
	}

	return embeddingsResponse{
		Object: "list",
		Data:   data,
		Model:  request.Model,
		Usage:  embeddingMockUsage,
	}
}

// 通用 embeddings 数据结构（兼容 OpenAI 格式）
type embeddingsRequest struct {
	Model          string      `json:"model" validate:"required"`
	Input          interface{} `json:"input" validate:"required"`
	EncodingFormat string      `json:"encoding_format,omitempty"`
	Dimensions     int         `json:"dimensions,omitempty"`
	User           string      `json:"user,omitempty"`
}

type embeddingsResponse struct {
	Object string      `json:"object"`
	Data   []embedding `json:"data"`
	Model  string      `json:"model"`
	Usage  usage       `json:"usage"`
}

type embedding struct {
	Object    string    `json:"object"`
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type usage struct {
	PromptTokens int `json:"prompt_tokens,omitempty"`
	TotalTokens  int `json:"total_tokens,omitempty"`
}
