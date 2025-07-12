package chat

import (
	"net/http"

	"llm-mock-server/pkg/provider"
	"llm-mock-server/pkg/utils"

	"github.com/gin-gonic/gin"
)

type requestHandler interface {
	provider.CommonRequestHandler

	HandleChatCompletions(context *gin.Context)
}

var (
	chatCompletionsHandlers = []requestHandler{
		&minimaxProvider{},
		&difyProvider{},
		&qwenProvider{},
		&openAiProvider{}, // As the last fallback
	}

	chatCompletionsRoutes = []string{
		// baidu
		"/v2/chat/completions",
		// doubao
		"/api/v3/chat/completions",
		// github
		"/chat/completions",
		// groq
		"/openai/v1/chat/completions",
		// minimax
		"/v1/text/chatcompletion_v2",
		"/v1/text/chatcompletion_pro",
		// openai
		"/v1/chat/completions",
		// qwen
		"/compatible-mode/v1/chat/completions",
		"/api/v1/services/aigc/text-generation/generation",
		// zhipu
		"/api/paas/v4/chat/completions",
		// dify
		"/v1/completion-messages",
		"/v1/chat-messages",
	}
)

func SetupRoutes(server *gin.Engine) {
	for _, route := range chatCompletionsRoutes {
		server.POST(route, handleChatCompletions)
	}
}

func handleChatCompletions(context *gin.Context) {
	if err := utils.BuildRequestContext(context); err != nil {
		return
	}

	for _, handler := range chatCompletionsHandlers {
		if handler.ShouldHandleRequest(context) {
			handler.HandleChatCompletions(context)
			return
		}
	}
	context.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
}
