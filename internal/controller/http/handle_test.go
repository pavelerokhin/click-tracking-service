package http

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"rsclabs-test/internal/repository"
	"rsclabs-test/internal/repository/inmemorystorage"
	"rsclabs-test/internal/service"
	"rsclabs-test/pkg/observe"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestRoutes() *routes {
	storage := inmemorystorage.NewInMemoryStorage(100, nil)
	bannerRepo, _ := repository.NewBannerRepository(storage)
	app := fiber.New()
	logger := observe.NewZapLogger("test-app")
	statsService := service.NewStatisticsService(bannerRepo, app, logger)

	return &routes{
		banners:    bannerRepo,
		statistics: statsService,
		l:          logger,
	}
}

func TestHandleClick(t *testing.T) {
	app := fiber.New()
	routes := setupTestRoutes()

	app.Post("/click/:bannerID", routes.handleClick)

	tests := []struct {
		name           string
		bannerID       string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Valid banner ID",
			bannerID:       "1",
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"bannerID": float64(1),
				"success":  true,
			},
		},
		{
			name:           "Invalid banner ID format",
			bannerID:       "abc",
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "Invalid banner ID format",
			},
		},
		{
			name:           "Banner ID out of range",
			bannerID:       "101",
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "Banner ID must be between 1 and 100",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/click/"+tt.bannerID, nil)
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}

func TestHandleStatsRequest(t *testing.T) {
	app := fiber.New()
	routes := setupTestRoutes()

	app.Post("/stats/:bannerID", routes.handleStatsRequest)

	tests := []struct {
		name           string
		bannerID       string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:     "Valid request",
			bannerID: "1",
			requestBody: map[string]interface{}{
				"from": "2024-01-01",
				"to":   "2024-01-31",
			},
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"stats": []interface{}{},
			},
		},
		{
			name:           "Missing banner ID",
			bannerID:       "",
			requestBody:    map[string]interface{}{},
			expectedStatus: 404,
			expectedBody:   nil,
		},
		{
			name:     "Invalid banner ID format",
			bannerID: "abc",
			requestBody: map[string]interface{}{
				"from": "2024-01-01",
				"to":   "2024-01-31",
			},
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "Invalid banner ID format",
			},
		},
		{
			name:     "Invalid JSON body",
			bannerID: "1",
			requestBody: map[string]interface{}{
				"from": "invalid-date",
				"to":   "invalid-date",
			},
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"stats": []interface{}{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/stats/"+tt.bannerID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				json.NewDecoder(resp.Body).Decode(&response)
				assert.Equal(t, tt.expectedBody, response)
			} else {
				assert.Empty(t, resp.Body)
			}
		})
	}
}

func TestGetBannerID(t *testing.T) {
	app := fiber.New()

	app.Get("/:bannerID", func(c *fiber.Ctx) error {
		id, err := getBannerID(c)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.JSON(fiber.Map{"id": id})
	})

	tests := []struct {
		name           string
		bannerID       string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "Valid banner ID",
			bannerID:       "1",
			expectedStatus: 200,
			expectedBody: map[string]interface{}{
				"id": float64(1),
			},
		},
		{
			name:           "Missing banner ID",
			bannerID:       "",
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "banner ID is required",
			},
		},
		{
			name:           "Invalid banner ID format",
			bannerID:       "abc",
			expectedStatus: 400,
			expectedBody: map[string]interface{}{
				"error": "invalid banner ID format: strconv.Atoi: parsing \"abc\": invalid syntax",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/"+tt.bannerID, nil)
			resp, _ := app.Test(req)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			var response map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&response)
			assert.Equal(t, tt.expectedBody, response)
		})
	}
}
