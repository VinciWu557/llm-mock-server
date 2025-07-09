package embeddings

import (
	"errors"
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

// qwen 错误响应结构
type qwenErrorResp struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

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
		// 处理兼容模式请求
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

		// 构建 qwen 请求格式
		qwenRequest, err := h.buildQwenTextEmbeddingRequest(&compatRequest)
		if err != nil {
			h.sendErrorResponse(c, http.StatusBadRequest,
				"InvalidParameter", fmt.Sprintf("invalid input: %v", err.Error()))
			return
		}

		// 生成 qwen 响应并转换为兼容格式
		qwenResponse := h.createQwenTextEmbeddingResponse(*qwenRequest)
		response := h.buildEmbeddingsResponse(&compatRequest, &qwenResponse)
		c.JSON(http.StatusOK, response)
	}
}

// 发送错误响应
func (h *qwenEmbeddings) sendErrorResponse(ctx *gin.Context, statusCode int, errorCode, errorMsg string) {
	errorResp := qwenErrorResp{
		Code:      errorCode,
		Message:   errorMsg,
		RequestId: embeddingMockId,
	}
	ctx.JSON(statusCode, errorResp)
}

// 构建 qwen 文本嵌入请求
func (h *qwenEmbeddings) buildQwenTextEmbeddingRequest(request *embeddingsRequest) (*qwenTextEmbeddingRequest, error) {
	var texts []string
	if str, isString := request.Input.(string); isString {
		texts = []string{str}
	} else if strs, isArray := request.Input.([]interface{}); isArray {
		texts = make([]string, 0, len(strs))
		for _, item := range strs {
			if str, isString := item.(string); isString {
				texts = append(texts, str)
			} else {
				return nil, errors.New("unsupported input type in array: " + reflect.TypeOf(item).String())
			}
		}
	} else {
		return nil, errors.New("unsupported input type: " + reflect.TypeOf(request.Input).String())
	}
	return &qwenTextEmbeddingRequest{
		Model: request.Model,
		Input: qwenTextEmbeddingInput{
			Texts: texts,
		},
	}, nil
}

// 构建通用 embeddings 响应
func (h *qwenEmbeddings) buildEmbeddingsResponse(request *embeddingsRequest, qwenResponse *qwenTextEmbeddingResponse) *embeddingsResponse {
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

// 创建 qwen 文本嵌入响应
func (h *qwenEmbeddings) createQwenTextEmbeddingResponse(request qwenTextEmbeddingRequest) qwenTextEmbeddingResponse {
	embeddings := make([]qwenTextEmbeddings, len(request.Input.Texts))
	for i, _ := range request.Input.Texts {
		// 生成 mock embedding 向量 (1536 维度，模拟真实的 embedding)
		mockEmbedding := make([]float64, 1536)
		for j := range mockEmbedding {
			mockEmbedding[j] = 0.001 * float64(i+j) // 简单的 mock 数据
		}

		embeddings[i] = qwenTextEmbeddings{
			TextIndex: i,
			Embedding: mockEmbedding,
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
