package worker

import (
	"context"
	"rsclabs-test/internal/repository"
	"rsclabs-test/internal/repository/inmemorystorage"
	"rsclabs-test/internal/service"
	"rsclabs-test/pkg/observe"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestWorker() (*StatisticsWorker, context.CancelFunc) {
	storage := inmemorystorage.NewInMemoryStorage(100, nil)
	bannerRepo, _ := repository.NewBannerRepository(storage)
	app := fiber.New()
	logger := observe.NewZapLogger("test-app")
	statsService := service.NewStatisticsService(bannerRepo, app, logger)
	worker := NewStatisticsWorker(bannerRepo, statsService, logger)
	_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	return worker, cancel
}

func TestStatisticsWorkerRun(t *testing.T) {
	worker, cancel := setupTestWorker()
	defer cancel()

	// Start the worker
	worker.Run(context.Background())

	for i := 0; i < 5; i++ {
		worker.bannerRepository.RegisterClick(1)
	}

	// Wait for at least one statistics update
	time.Sleep(statisticsUpdateInterval + 100*time.Millisecond)

	// Check if statistics were registered
	snapshots := worker.statisticsService.GetSnapshots()
	assert.Greater(t, len(snapshots), 0, "Expected at least one snapshot to be created")

	// Verify the snapshot content
	lastSnapshot := snapshots[len(snapshots)-1]
	assert.NotEmpty(t, lastSnapshot.Banners, "Expected non-empty banners in snapshot")
}

func TestStatisticsWorkerShutdown(t *testing.T) {
	worker, cancel := setupTestWorker()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start the worker
	worker.Run(ctx)

	time.Sleep(200 * time.Millisecond)

	snapshots := worker.statisticsService.GetSnapshots()
	initialCount := len(snapshots)

	// Wait a bit more to ensure no more updates
	time.Sleep(statisticsUpdateInterval + 100*time.Millisecond)

	// Verify no new snapshots were created after shutdown
	assert.Equal(t, initialCount, len(worker.statisticsService.GetSnapshots()),
		"Expected no new snapshots after worker shutdown")
}

func TestStatisticsWorkerConcurrentClicks(t *testing.T) {
	worker, cancel := setupTestWorker()
	defer cancel()

	// Start the worker
	worker.Run(context.Background())

	// Simulate concurrent clicks
	iterations := 100
	goroutines := 10
	done := make(chan bool)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				worker.bannerRepository.RegisterClick(1)
			}
			done <- true
		}()
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Wait for statistics update
	time.Sleep(statisticsUpdateInterval + 100*time.Millisecond)

	// Verify statistics were registered correctly
	snapshots := worker.statisticsService.GetSnapshots()
	assert.Greater(t, len(snapshots), 0, "Expected at least one snapshot to be created")

	// Verify the total count in the last snapshot
	lastSnapshot := snapshots[len(snapshots)-1]
	found := false
	for _, banner := range lastSnapshot.Banners {
		if banner.BannerID == 1 {
			found = true
			expectedCount := goroutines * iterations
			assert.Equal(t, expectedCount, banner.Count,
				"Expected count %d, got %d", expectedCount, banner.Count)
			break
		}
	}
	assert.True(t, found, "Banner not found in snapshot")
}

func TestStatisticsWorkerEmptySnapshots(t *testing.T) {
	worker, cancel := setupTestWorker()
	defer cancel()

	// Start the worker
	worker.Run(context.Background())

	// Wait for statistics update
	time.Sleep(statisticsUpdateInterval + 100*time.Millisecond)

	// Verify no snapshots were created when there are no clicks
	snapshots := worker.statisticsService.GetSnapshots()
	assert.Equal(t, 0, len(snapshots), "Expected no snapshots when there are no clicks")
}
