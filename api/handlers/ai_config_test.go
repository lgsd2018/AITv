package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

type apiResponse struct {
	Success bool                `json:"success"`
	Data    json.RawMessage     `json:"data"`
	Error   *response.ErrorInfo `json:"error"`
}

func setupAIConfigHandler(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        "file::memory:?cache=shared",
	}, &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	if err := db.AutoMigrate(&models.AIServiceConfig{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	cfg := &config.Config{
		Storage: config.StorageConfig{
			LocalPath: "",
			BaseURL:   "",
		},
	}
	log := logger.NewLogger(true)
	handler := NewAIConfigHandler(db, cfg, log)

	r := gin.New()
	r.GET("/api/v1/ai-configs", handler.ListConfigs)
	return r, db
}

func TestListConfigsInvalidServiceType(t *testing.T) {
	r, _ := setupAIConfigHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai-configs?service_type=bad", nil)
	recorder := httptest.NewRecorder()

	start := time.Now()
	r.ServeHTTP(recorder, req)
	elapsed := time.Since(start)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("response time exceeded 500ms: %v", elapsed)
	}

	var resp apiResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected success false")
	}
	if resp.Error == nil || resp.Error.Code != "BAD_REQUEST" {
		t.Fatalf("expected BAD_REQUEST error")
	}
}

func TestListConfigsSuccess(t *testing.T) {
	r, db := setupAIConfigHandler(t)
	err := db.Create(&models.AIServiceConfig{
		ServiceType: "text",
		Name:        "test",
		BaseURL:     "http://example.com",
		APIKey:      "key",
		Model:       models.ModelField{"model"},
		Priority:    0,
		IsActive:    true,
	}).Error
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai-configs?service_type=text", nil)
	recorder := httptest.NewRecorder()

	start := time.Now()
	r.ServeHTTP(recorder, req)
	elapsed := time.Since(start)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("response time exceeded 500ms: %v", elapsed)
	}

	var resp apiResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success true")
	}

	var configs []models.AIServiceConfig
	if err := json.Unmarshal(resp.Data, &configs); err != nil {
		t.Fatalf("failed to decode configs: %v", err)
	}
	if len(configs) != 1 {
		t.Fatalf("expected 1 config, got %d", len(configs))
	}
}

func TestListConfigsDatabaseNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Storage: config.StorageConfig{
			LocalPath: "",
			BaseURL:   "",
		},
	}
	log := logger.NewLogger(true)
	handler := NewAIConfigHandler(nil, cfg, log)

	r := gin.New()
	r.GET("/api/v1/ai-configs", handler.ListConfigs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai-configs?service_type=text", nil)
	recorder := httptest.NewRecorder()

	start := time.Now()
	r.ServeHTTP(recorder, req)
	elapsed := time.Since(start)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", recorder.Code)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("response time exceeded 500ms: %v", elapsed)
	}

	var resp apiResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected success false")
	}
	if resp.Error == nil || resp.Error.Code != "INTERNAL_ERROR" {
		t.Fatalf("expected INTERNAL_ERROR error")
	}
}
