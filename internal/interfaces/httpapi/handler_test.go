package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aroma-maintenance/internal/application"
)

func TestProductBrowseAndStockEndpoint(t *testing.T) {
	handler := NewHandler(application.NewMaintenanceService())
	page := httptest.NewRecorder()
	handler.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "琥珀木芯蜡烛") {
		t.Fatalf("unexpected page response: %d", page.Code)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/products/candle-amber/stock", strings.NewReader(`{"delta":2,"reason":"restock"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"stock":20`) {
		t.Fatalf("unexpected API response: %d %s", response.Code, response.Body.String())
	}
}
