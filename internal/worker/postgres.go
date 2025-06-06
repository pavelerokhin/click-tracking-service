package worker

import (
	"rsclabs-test/internal/repository"
	"rsclabs-test/internal/service"
	"rsclabs-test/pkg/observe"
)

type PostgresWorker struct {
	bannerRepository  *repository.BannerRepositoryInMemory
	statisticsService *service.StatisticsService
	l                 *observe.Logger
}

func NewPostgresWorker(
	bannerRepository *repository.BannerRepositoryInMemory,
	statistics *service.StatisticsService,
	l *observe.Logger,
) *PostgresWorker {
	return &PostgresWorker{
		bannerRepository:  bannerRepository,
		statisticsService: statistics,
		l:                 l,
	}
}
