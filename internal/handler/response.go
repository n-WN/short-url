package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// SuccessResponse 成功响应结构
type SuccessResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// respondWithError 返回错误响应
func respondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	})
}

// respondWithSuccess 返回成功响应
func respondWithSuccess(c *gin.Context, code int, data interface{}, message ...string) {
	response := SuccessResponse{Data: data}
	if len(message) > 0 {
		response.Message = message[0]
	}
	c.JSON(code, response)
}

// Health 健康检查响应
type Health struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Services  map[string]interface{} `json:"services,omitempty"`
}
