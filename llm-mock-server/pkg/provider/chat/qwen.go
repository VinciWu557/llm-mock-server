package chat

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"llm-mock-server/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

const (
	qwenDomain                      = "dashscope.aliyuncs.com"
	qwenChatCompletionPath          = "/api/v1/services/aigc/text-generation/generation"
	

	qwenCompatibleChatCompletionPath = "/compatible-mode/v1/chat/completions"
	qwenCompatibleCompletionsPath    = "/compatible-mode/v1/completions"

	qwenCompatibleFilesPath               = "/compatible-mode/v1/files"
	qwenCompatibleRetrieveFilePath        = "/compatible-mode/v1/files/{file_id}"
	qwenCompatibleRetrieveFileContentPath = "/compatible-mode/v1/files/{file_id}/content"
	qwenCompatibleBatchesPath             = "/compatible-mode/v1/batches"
	qwenCompatibleRetrieveBatchPath       = "/compatible-mode/v1/batches/{batch_id}"
	qwenBailianPath                       = "/api/v1/apps"
	qwenMultimodalGenerationPath          = "/api/v1/services/aigc/multimodal-generation/generation"
	qwenResultFormatMessage               = "message"
)

type qwenProvider struct {
}

func (p *qwenProvider) ShouldHandleRequest(ctx *gin.Context) bool {
	paths := []string{
		qwenChatCompletionPath,
		qwenCompatibleChatCompletionPath,
	}

	context, _ := utils.GetRequestContext(ctx)
	if context.Host == qwenDomain && slices.Contains(paths, context.Path) {
		return true
	}
	return false
}

func (p *qwenProvider) HandleChatCompletions(ctx *gin.Context) {
	// Validate Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		p.sendErrorResponse(ctx, http.StatusUnauthorized,
			"InvalidApiKey", "No API-key provided.")
		return
	}

	// Determine if the request is a stream request
	isStream := p.isStreamRequest(ctx)

	// 根据不同路径处理不同类型的请求
	switch ctx.Request.URL.Path {
	case qwenChatCompletionPath:
		// 处理原生 Qwen 请求
		var qwenRequest qwenTextGenRequest
		if err := ctx.ShouldBindJSON(&qwenRequest); err != nil {
			p.sendErrorResponse(ctx, http.StatusBadRequest,
				"InvalidParameter", fmt.Sprintf("invalid params: %v", err.Error()))
			return
		}

		// Validate request body
		if err := utils.Validate.Struct(qwenRequest); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			for _, fieldError := range validationErrors {
				p.sendErrorResponse(ctx, http.StatusBadRequest,
					"InvalidParameter", fmt.Sprintf("invalid params: %v", fieldError.Error()))
				return
			}
		}

		prompt := ""
		messages := qwenRequest.Input.Messages
		if len(messages) > 0 && messages[len(messages)-1].IsStringContent() {
			prompt = messages[len(messages)-1].StringContent()
		}
		response := prompt2Response(prompt)

		if isStream {
			p.handleStreamResponse(ctx, qwenRequest, response)
		} else {
			p.handleNonStreamResponse(ctx, qwenRequest, response)
		}
	case qwenCompatibleChatCompletionPath:
		// 处理兼容模式请求
		var compatRequest chatCompletionRequest
		if err := ctx.ShouldBindJSON(&compatRequest); err != nil {
			p.sendErrorResponse(ctx, http.StatusBadRequest,
				"InvalidParameter", fmt.Sprintf("invalid params: %v", err.Error()))
			return
		}

		// Validate request body
		if err := utils.Validate.Struct(compatRequest); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			for _, fieldError := range validationErrors {
				p.sendErrorResponse(ctx, http.StatusBadRequest,
					"InvalidParameter", fmt.Sprintf("invalid params: %v", fieldError.Error()))
				return
			}
		}

		prompt := ""
		if len(compatRequest.Messages) > 0 {
			prompt = extractPromptFromMessages(compatRequest.Messages)
		}
		response := prompt2Response(prompt)

		if isStream {
			// 实现流式响应，参考 openAiProvider 的实现
			utils.SetEventStreamHeaders(ctx)
			dataChan := make(chan string)
			stopChan := make(chan bool, 1)
			streamResponse := chatCompletionResponse{
				Id:      completionMockId,
				Object:  objectChatCompletionChunk,
				Created: completionMockCreated,
				Model:   compatRequest.Model,
			}
			streamResponseChoice := chatCompletionChoice{Delta: &chatMessage{}}

			go func() {
				for i, s := range response {
					streamResponseChoice.Delta.Content = string(s)
					if i == len(response)-1 {
						streamResponseChoice.FinishReason = ptr(stopReason)
					}
					streamResponse.Choices = []chatCompletionChoice{streamResponseChoice}
					jsonStr, _ := json.Marshal(streamResponse)
					dataChan <- string(jsonStr)

					// 模拟响应延迟
					time.Sleep(200 * time.Millisecond)
				}
				stopChan <- true
			}()

			ctx.Stream(func(w io.Writer) bool {
				select {
				case data := <-dataChan:
					ctx.Render(-1, streamEvent{Data: "data: " + data})
					return true
				case <-stopChan:
					ctx.Render(-1, streamEvent{Data: "data: [DONE]"})
					return false
				}
			})
		} else {
			// 使用与 OpenAI 相同的响应格式
			completion := createChatCompletionResponse(compatRequest.Model, response)
			ctx.JSON(http.StatusOK, completion)
		}
	}
}

// 从兼容模式消息中提取提示文本
func extractPromptFromMessages(messages []chatMessage) string {
	if len(messages) == 0 {
		return ""
	}
	lastMessage := messages[len(messages)-1]
	return lastMessage.StringContent()
}

func (p *qwenProvider) sendErrorResponse(ctx *gin.Context, statusCode int, errorCode, errorMsg string) {
	errorResp := qwenErrorResp{
		Code:      errorCode,
		Message:   errorMsg,
		RequestId: completionMockId,
	}
	ctx.JSON(statusCode, errorResp)
}

// isStreamRequest checks if the request is a stream request.
func (p *qwenProvider) isStreamRequest(ctx *gin.Context) bool {
	acceptHeader := ctx.GetHeader("Accept")
	sseHeader := ctx.GetHeader("X-DashScope-SSE")

	// Check if Accept header is text/event-stream or X-DashScope-SSE is set to enable
	if acceptHeader == "text/event-stream" || sseHeader == "enable" {
		return true
	}
	return false
}

func (p *qwenProvider) handleNonStreamResponse(ctx *gin.Context, chatRequest qwenTextGenRequest, response string) {
	completion := createQwenTextGenResponse(chatRequest, response)
	ctx.JSON(http.StatusOK, completion)
}

func (p *qwenProvider) handleStreamResponse(ctx *gin.Context, chatRequest qwenTextGenRequest, response string) {
	utils.SetEventStreamHeaders(ctx)
	dataChan, stopChan := createQwenStreamResponse(chatRequest, response)
	ctx.Stream(func(w io.Writer) bool {
		select {
		case data := <-dataChan:
			ctx.Render(-1, streamEvent{Data: "data: " + data})
			return true
		case <-stopChan:
			ctx.Render(-1, streamEvent{Data: "data: [DONE]"})
			return false
		}
	})
}

type qwenErrorResp struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestId string `json:"request_id"`
}

type qwenTextGenRequest struct {
	Model      string                `json:"model"`
	Input      qwenTextGenInput      `json:"input"`
	Parameters qwenTextGenParameters `json:"parameters,omitempty"`
}

type qwenTextGenInput struct {
	Messages []qwenMessage `json:"messages"`
}

type qwenTextGenParameters struct {
	ResultFormat      string  `json:"result_format,omitempty"`
	MaxTokens         int     `json:"max_tokens,omitempty"`
	RepetitionPenalty float64 `json:"repetition_penalty,omitempty"`
	N                 int     `json:"n,omitempty"`
	Seed              int     `json:"seed,omitempty"`
	Temperature       float64 `json:"temperature,omitempty"`
	TopP              float64 `json:"top_p,omitempty"`
	IncrementalOutput bool    `json:"incremental_output,omitempty"`
	EnableSearch      bool    `json:"enable_search,omitempty"`
	Tools             []tool  `json:"tools,omitempty"`
}

type qwenTextGenResponse struct {
	RequestId string            `json:"request_id"`
	Output    qwenTextGenOutput `json:"output"`
	Usage     qwenUsage         `json:"usage"`
}

func createQwenTextGenResponse(chatRequest qwenTextGenRequest, response string) qwenTextGenResponse {
	var output qwenTextGenOutput
	if chatRequest.Parameters.ResultFormat == qwenResultFormatMessage {
		output = qwenTextGenOutput{
			Choices: []qwenTextGenChoice{
				{
					FinishReason: stopReason,
					Message: qwenMessage{
						Role:    roleAssistant,
						Content: response,
					},
				},
			},
		}
	} else {
		output = qwenTextGenOutput{
			FinishReason: stopReason,
			Text:         response,
		}
	}
	return qwenTextGenResponse{
		Output: output,
		Usage: qwenUsage{
			InputTokens:  9,
			OutputTokens: 1,
			TotalTokens:  10,
		},
		RequestId: completionMockId,
	}
}

func createQwenStreamResponse(chatRequest qwenTextGenRequest, response string) (chan string, chan bool) {
	dataChan := make(chan string)
	stopChan := make(chan bool, 1)
	streamResponse := chatCompletionResponse{
		Id:      completionMockId,
		Object:  objectChatCompletionChunk,
		Created: completionMockCreated,
		Model:   chatRequest.Model,
	}
	streamResponseChoice := chatCompletionChoice{Delta: &chatMessage{}}

	go func() {
		for i, s := range response {
			streamResponseChoice.Delta.Content = string(s)
			if i == len(response)-1 {
				streamResponseChoice.FinishReason = ptr(stopReason)
			}
			streamResponse.Choices = []chatCompletionChoice{streamResponseChoice}
			jsonStr, _ := json.Marshal(streamResponse)
			dataChan <- string(jsonStr)

			// 模拟响应延迟
			time.Sleep(200 * time.Millisecond)
		}
		stopChan <- true
	}()

	return dataChan, stopChan
}

type qwenTextGenOutput struct {
	FinishReason string              `json:"finish_reason,omitempty"`
	Text         string              `json:"text,omitempty"`
	Choices      []qwenTextGenChoice `json:"choices,omitempty"`
}

type qwenTextGenChoice struct {
	FinishReason string      `json:"finish_reason"`
	Message      qwenMessage `json:"message"`
}

type qwenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type qwenMessage struct {
	Name      string     `json:"name,omitempty"`
	Role      string     `json:"role"`
	Content   any        `json:"content"`
	ToolCalls []toolCall `json:"tool_calls,omitempty"`
}

func (m *qwenMessage) IsStringContent() bool {
	_, ok := m.Content.(string)
	return ok
}

func (m *qwenMessage) StringContent() string {
	content, ok := m.Content.(string)
	if ok {
		return content
	}
	contentList, ok := m.Content.([]any)
	if ok {
		var contentStr string
		for _, contentItem := range contentList {
			contentMap, ok := contentItem.(map[string]any)
			if !ok {
				continue
			}
			if contentMap["type"] == contentTypeText {
				if subStr, ok := contentMap[contentTypeText].(string); ok {
					contentStr += subStr + "\n"
				}
			}
		}
		return contentStr
	}
	return ""
}
