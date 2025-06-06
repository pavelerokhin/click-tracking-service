package http

import (
	"fmt"
	"rsclabs-test/internal/model"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"rsclabs-test/internal/repository"
	"rsclabs-test/internal/service"
	"rsclabs-test/pkg/observe"
)

type routes struct {
	banners    *repository.BannerRepositoryInMemory
	statistics *service.StatisticsService
	l          *observe.Logger
}

type BannerStorage interface {
	RegisterClick(bannerID int) bool
	GetCountSnapshot() model.Snapshot
}

func (r *routes) handleClick(c *fiber.Ctx) error {
	bannerID := c.Params("bannerID")

	bid, err := strconv.Atoi(bannerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid banner ID format",
		})
	}

	if bid < 1 || bid > r.banners.MaxBanners {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Banner ID must be between 1 and %d", r.banners.MaxBanners),
		})
	}

	err = r.banners.RegisterClick(bid - 1)
	if err != nil {
		r.l.Error(fmt.Errorf("failed to register click: %w", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to register click",
		})
	}

	return c.JSON(fiber.Map{
		"bannerID": bid,
		"success":  true,
	})
}

func (r *routes) handleStatsRequest(c *fiber.Ctx) error {
	bannerID := c.Params("bannerID")
	if bannerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Banner ID is required"})
	}

	bid, err := strconv.Atoi(bannerID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid banner ID format"})
	}

	// Parse JSON body
	var requestBody struct {
		From string `json:"from"`
		To   string `json:"to"`
	}

	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON body"})
	}

	request := model.StatisticsRequest{
		BannerID: bid,
		From:     requestBody.From,
		To:       requestBody.To,
	}

	stats, err := r.statistics.GetStatistics(c.Context(), request)
	if err != nil {
		// Check if it's a "no data" error specifically
		if strings.Contains(err.Error(), "no statistics data available") {
			// Return empty stats instead of error
			return c.JSON(fiber.Map{
				"stats": []interface{}{}, // Empty array
			})
		}

		// For other errors, log and return 500
		r.l.Error(fmt.Errorf("failed to get statistics: %w", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to retrieve statistics"})
	}

	// Check if stats is empty using your existing method
	if stats.IsEmpty() {
		r.l.Debug("no statistics found for banner ID", map[string]any{
			"banner_id": bannerID,
		})
		// Return empty stats instead of 404
		return c.JSON(fiber.Map{
			"stats": []interface{}{}, // Empty array
		})
	}

	r.l.Debug("statistics retrieved successfully", map[string]any{"stats": stats})
	return c.JSON(stats)
}

func getBannerID(c *fiber.Ctx) (int, error) {
	bannerID := c.Params("bannerID")
	if bannerID == "" {
		return 0, fmt.Errorf("banner ID is required")
	}

	bid, err := strconv.Atoi(bannerID)
	if err != nil {
		return 0, fmt.Errorf("invalid banner ID format: %w", err)
	}

	return bid, nil
}
