package member

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	service "github.com/rafa/golang-cc/internal/services/member"
)

func TestCreateCardRejectsInvalidRequestBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.POST("/cards", New(&service.Service{}).CreateCard)
	request := httptest.NewRequest(http.MethodPost, "/cards", strings.NewReader(`{}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestCreateTransactionRejectsInvalidDateBeforeService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.POST("/transactions", New(&service.Service{}).CreateTransaction)
	body := `{"card_id":"card-id","amount_minor":100,"category_id":"dining","merchant_name":"store","transaction_date":"bad"}`
	request := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
