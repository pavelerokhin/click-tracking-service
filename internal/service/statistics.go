package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"rsclabs-test/internal/model"
	"rsclabs-test/internal/repository"
	"rsclabs-test/pkg/observe"
)

const (
	defaultTimeout = 10 * time.Second
)

type StatisticsService struct {
	bannerRepo *repository.BannerRepositoryInMemory
	snapshots  []model.Snapshot
	server     *fiber.App
	l          *observe.Logger
}

func NewStatisticsService(
	repo *repository.BannerRepositoryInMemory,
	hs *fiber.App,
	l *observe.Logger,
) *StatisticsService {
	return &StatisticsService{
		bannerRepo: repo,
		snapshots:  make([]model.Snapshot, 0),
		server:     hs,
		l:          l,
	}
}

func (s *StatisticsService) RegisterStatistics(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	cs := s.bannerRepo.GetCountSnapshot()
	if cs.IsEmpty() {
		s.l.Debug("no new statistics data to update")
		return
	}

	cs.TimeStamp = time.Now()

	s.l.Debug("*** registering new statistics snapshot ***", map[string]any{
		"snapshot": cs,
	})

	s.snapshots = append(s.snapshots, cs)

	s.bannerRepo.ZeroOutCounts()
}

func (s *StatisticsService) GetStatistics(
	ctx context.Context,
	request model.StatisticsRequest,
) (model.StatisticsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	if len(s.snapshots) == 0 {
		return model.StatisticsResponse{}, nil
	}

	if request.BannerID < 1 || request.BannerID > s.bannerRepo.MaxBanners {
		return model.StatisticsResponse{}, fmt.Errorf("invalid banner id %d", request.BannerID)
	}

	from, err := s.getFrom(request)
	if err != nil {
		return model.StatisticsResponse{}, fmt.Errorf("failed to parse 'from' time: %w", err)
	}

	to, err := s.getTo(request)
	if err != nil {
		return model.StatisticsResponse{}, fmt.Errorf("failed to parse 'to' time: %w", err)
	}

	if from.After(to) {
		return model.StatisticsResponse{}, fmt.Errorf("invalid time range: from %s is after to %s", from, to)
	}

	out := model.StatisticsResponse{
		Stats: make([]model.Banner, 0),
	}
	for _, snapshot := range s.snapshots {
		if snapshot.TimeStamp.Before(from) || snapshot.TimeStamp.After(to) {
			continue
		}

		filtered, ok := snapshot.Banners[request.BannerID-1]
		if !ok {
			continue
		}
		filtered.TimeStamp = snapshot.TimeStamp

		out.Stats = append(out.Stats, filtered)
	}

	return out, nil
}

func (s *StatisticsService) GetSnapshots() []model.Snapshot {
	return s.snapshots
}

func (s *StatisticsService) getFrom(request model.StatisticsRequest) (time.Time, error) {
	if request.From != "" {
		// Parse as local time, then convert to UTC to match storage
		parsedFrom, err := time.ParseInLocation("2006-01-02T15:04:05", request.From, time.Local)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid 'from' time format: %w", err)
		}

		// Convert to UTC to match how you store clicks
		parsedFromUTC := parsedFrom.UTC()

		s.l.Debug(fmt.Sprintf("*** parsing 'from' time: input=%s, local=%s, UTC=%s",
			request.From,
			parsedFrom.Format("2006-01-02T15:04:05"),
			parsedFromUTC.Format("2006-01-02T15:04:05")))

		return parsedFromUTC, nil
	}

	return time.Time{}, nil
}

func (s *StatisticsService) getTo(request model.StatisticsRequest) (time.Time, error) {
	if request.To != "" {
		parsedTo, err := time.ParseInLocation("2006-01-02T15:04:05", request.To, time.Local)
		if err != nil {
			return time.Now(), err
		}

		parsedToUTC := parsedTo.UTC()

		s.l.Debug(fmt.Sprintf("*** parsing 'to' time: input=%s, local=%s, UTC=%s",
			request.To,
			parsedTo.Format("2006-01-02T15:04:05"),
			parsedToUTC.Format("2006-01-02T15:04:05")))

		return parsedToUTC, nil
	}

	return time.Now(), nil
}
