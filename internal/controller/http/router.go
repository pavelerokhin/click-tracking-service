package http

import (
	"github.com/gofiber/fiber/v2"
	"rsclabs-test/internal/service"

	"rsclabs-test/internal/repository"
	"rsclabs-test/pkg/observe"
)

func NewRouter(
	banners *repository.BannerRepositoryInMemory,
	statisticsService *service.StatisticsService,
	s *fiber.App,
	l *observe.Logger,
) {
	r := &routes{
		banners:    banners,
		statistics: statisticsService,
		l:          l,
	}
	s.Get("/counter/:bannerID", r.handleClick)

	s.Post("/stats/:bannerID", r.handleStatsRequest)
}
