package utils

import (
	"fmt"

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
