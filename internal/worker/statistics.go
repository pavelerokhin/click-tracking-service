package worker

import (
	"context"
	"time"

	"rsclabs-test/internal/repository"
	"rsclabs-test/internal/service"
	"rsclabs-test/pkg/observe"
)

const (
	statisticsUpdateInterval = 1 * time.Minute
)

type StatisticsWorker struct {
	bannerRepository  *repository.BannerRepositoryInMemory
	statisticsService *service.StatisticsService
	l                 *observe.Logger
}

func NewStatisticsWorker(
	bannerRepository *repository.BannerRepositoryInMemory,
	statistics *service.StatisticsService,
	l *observe.Logger,
) *StatisticsWorker {
	return &StatisticsWorker{
		bannerRepository:  bannerRepository,
		statisticsService: statistics,
		l:                 l,
	}
}

func (w *StatisticsWorker) Run(ctx context.Context) {
	w.l.Info("starting statisticsService service with poll", map[string]interface{}{"interval": statisticsUpdateInterval})

	go func() {
		timer := time.NewTimer(0)
		for {
			select {
			case <-timer.C:
				w.l.Debug("updating statisticsService", map[string]any{"len snapshots now": len(w.statisticsService.GetSnapshots())})

				w.statisticsService.RegisterStatistics(ctx)

				timer.Reset(statisticsUpdateInterval)
			case <-ctx.Done(): // exit
				w.l.Info("stopping statisticsService worker")

				return
			}
		}
	}()
}
