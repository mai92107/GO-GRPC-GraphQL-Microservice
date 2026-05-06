package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ResponseBase[T any] struct{
	Code          int         `json:"code"`
	Message       string      `json:"message"`
	Data          T			  `json:"data"`
	RequestedTime time.Time   `json:"requestedTime"`
	RespondedTime time.Time   `json:"respondedTime"`
}

// 成功回傳 (HTTP 200)
func Success[T any](c *gin.Context, requestTime time.Time, data T) {
	c.JSON(http.StatusOK, ResponseBase[T]{
		Code:          200,
		Message:       "OK",
		Data:          data,
		RequestedTime: requestTime,
		RespondedTime: time.Now().UTC(),
	})
}

// 錯誤回傳 (可自訂 HTTP status 與錯誤代碼)
func Error(c *gin.Context, code int, requestTime time.Time, msg string) {
	c.JSON(http.StatusOK, ResponseBase[any]{
		Code:          code,
		Message:       msg,
		Data:          nil,
		RequestedTime: requestTime,
		RespondedTime: time.Now().UTC(),
	})
}