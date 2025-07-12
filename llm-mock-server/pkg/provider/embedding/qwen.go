package embedding

import (
	"fmt"
	"net/http"
	"reflect"

	"llm-mock-server/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	qwenDomain                      = "dashscope.aliyuncs.com"
	qwenTextEmbeddingPath           = "/api/v1/services/embeddings/text-embedding/text-embedding"
	qwenCompatibleTextEmbeddingPath = "/compatible-mode/v1/embeddings"

	// Mock constants
	embeddingMockId = "embedding-mock"
)

var (
	// 固定的 mock embedding 向量，用于测试
	qwenMockEmbeddingVector = []float64{0.001, 0.002, 0.003}

	// 固定的 usage 数据
	qwenMockUsage = usage{
		PromptTokens: 5,
		TotalTokens:  5,
	}
)

// qwen 错误响应结构
type qwenErrorResp struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

type qwenEmbedding struct {
}

func (h *qwenEmbedding) ShouldHandleRequest(ctx *gin.Context) bool {
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

func (h *qwenEmbedding) HandleEmbeddings(c *gin.Context) {
	// 验证 Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.sendErrorResponse(c, http.StatusUnauthorized,
			"InvalidApiKey", "No API-key provided.")
		return
	}

	// 根据不同路径处理不同类型的请求
	switch c.Request.URL.Path {
	case qwenTextEmbeddingPath:
		// 处理原生 Qwen embeddings 请求
		var qwenRequest qwenTextEmbeddingRequest
		if err := c.ShouldBindJSON(&qwenRequest); err != nil {
			h.sendErrorResponse(c, http.StatusBadRequest,
				"InvalidParameter", fmt.Sprintf("invalid params: %v", err.Error()))
			return
		}

		// 验证请求体
		if err := utils.Validate.Struct(qwenRequest); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			for _, fieldError := range validationErrors {
				h.sendErrorResponse(c, http.StatusBadRequest,
					"InvalidParameter", fmt.Sprintf("invalid params: %v", fieldError.Error()))
				return
			}
		}

		// 生成 mock embeddings 响应
		response := h.createQwenTextEmbeddingResponse(qwenRequest)
		c.JSON(http.StatusOK, response)

	case qwenCompatibleTextEmbeddingPath:
		// 处理兼容模式请求 - 直接处理，不经过 qwen 格式转换
		var compatRequest embeddingsRequest
		if err := c.ShouldBindJSON(&compatRequest); err != nil {
			h.sendErrorResponse(c, http.StatusBadRequest,
				"InvalidParameter", fmt.Sprintf("invalid params: %v", err.Error()))
			return
		}

		// 验证请求体
		if err := utils.Validate.Struct(compatRequest); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			for _, fieldError := range validationErrors {
				h.sendErrorResponse(c, http.StatusBadRequest,
					"InvalidParameter", fmt.Sprintf("invalid params: %v", fieldError.Error()))
				return
			}
		}

		// 直接处理输入文本
		texts, err := h.extractTextsFromInput(compatRequest.Input)
		if err != nil {
			h.sendErrorResponse(c, http.StatusBadRequest,
				"InvalidParameter", fmt.Sprintf("invalid input: %v", err.Error()))
			return
		}

		// 直接生成兼容格式的响应
		response := h.createCompatibleEmbeddingsResponse(compatRequest, texts)
		c.JSON(http.StatusOK, response)
	}
}

// 发送错误响应
func (h *qwenEmbedding) sendErrorResponse(ctx *gin.Context, statusCode int, errorCode, errorMsg string) {
	errorResp := qwenErrorResp{
		Code:      errorCode,
		Message:   errorMsg,
		RequestId: embeddingMockId,
	}
	ctx.JSON(statusCode, errorResp)
}

// 从输入中提取文本数组（类似 openai.go 的逻辑）
func (h *qwenEmbedding) extractTextsFromInput(input interface{}) ([]string, error) {
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

// 创建兼容格式的 embeddings 响应（类似 openai.go 的逻辑）
func (h *qwenEmbedding) createCompatibleEmbeddingsResponse(request embeddingsRequest, texts []string) embeddingsResponse {
	data := make([]embedding, len(texts))

	// 对每个文本都使用相同的固定 mock 向量
	for i := range texts {
		data[i] = embedding{
			Object:    "embedding",
			Index:     i,
			Embedding: qwenMockEmbeddingVector,
		}
	}

	return embeddingsResponse{
		Object: "list",
		Data:   data,
		Model:  request.Model,
		Usage:  qwenMockUsage,
	}
}

// 构建通用 embeddings 响应
func (h *qwenEmbedding) buildEmbeddingsResponse(request *embeddingsRequest, qwenResponse *qwenTextEmbeddingResponse) *embeddingsResponse {
	data := make([]embedding, 0, len(qwenResponse.Output.Embeddings))
	for _, qwenEmbedding := range qwenResponse.Output.Embeddings {
		data = append(data, embedding{
			Object:    "embedding",
			Index:     qwenEmbedding.TextIndex,
			Embedding: qwenEmbedding.Embedding,
		})
	}
	return &embeddingsResponse{
		Object: "list",
		Data:   data,
		Model:  request.Model,
		Usage: usage{
			PromptTokens: qwenResponse.Usage.TotalTokens,
			TotalTokens:  qwenResponse.Usage.TotalTokens,
		},
	}
}

// 创建 qwen 文本嵌入响应 - 使用固定值简化逻辑
func (h *qwenEmbedding) createQwenTextEmbeddingResponse(request qwenTextEmbeddingRequest) qwenTextEmbeddingResponse {
	embeddings := make([]qwenTextEmbeddings, len(request.Input.Texts))

	// 对每个文本都使用相同的固定 mock 向量
	for i := range request.Input.Texts {
		embeddings[i] = qwenTextEmbeddings{
			TextIndex: i,
			Embedding: qwenMockEmbeddingVector,
		}
	}

	return qwenTextEmbeddingResponse{
		RequestId: embeddingMockId,
		Output: qwenTextEmbeddingOutput{
			RequestId:  embeddingMockId,
			Embeddings: embeddings,
		},
		Usage: qwenUsage{
			InputTokens:  5,
			OutputTokens: 0,
			TotalTokens:  5,
		},
	}
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
