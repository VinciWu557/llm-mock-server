package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"llm-mock-server/pkg/log"

	"github.com/gin-gonic/gin"
)

type RequestContext struct {
	Host  string
	Path  string
	Model string
}

func GetRequestContext(context *gin.Context) (RequestContext, error) {
	requestCtx, exists := context.Get("requestContext")
	if !exists {
		return RequestContext{}, fmt.Errorf("request context not found")
	}

	ctx, ok := requestCtx.(RequestContext)
	if !ok {
		return RequestContext{}, fmt.Errorf("invalid request context type")
	}

	return ctx, nil
}

func BuildRequestContext(context *gin.Context) error {
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		log.Errorf("Error reading request body:", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Error reading request body"})
		return err
	}

	// Reset the request body so it can be read again by subsequent handlers
	context.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Errorf("Error unmarshalling JSON:", err)
		context.JSON(http.StatusBadRequest, gin.H{"error": "Error unmarshalling JSON"})
		return err
	}
	model, _ := data["model"].(string)

	context.Set("requestContext", RequestContext{
		Host:  context.Request.Host,
		Path:  context.Request.URL.Path,
		Model: model})

	return nil
}
