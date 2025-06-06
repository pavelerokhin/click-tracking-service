package repository

import (
	"testing"

	"rsclabs-test/internal/repository/inmemorystorage"
)

func setupTestRepository() *BannerRepositoryInMemory {
	storage := inmemorystorage.NewInMemoryStorage(100, nil) // Set max capacity to 100
	repo, _ := NewBannerRepository(storage)
	return repo
}

func TestRegisterClick(t *testing.T) {
	repo := setupTestRepository()

	// Test registering a click for a valid ID
	err := repo.RegisterClick(1)
	if err != nil {
		t.Errorf("RegisterClick failed: %v", err)
	}

	// Test registering multiple clicks
	for i := 0; i < 5; i++ {
		err = repo.RegisterClick(1)
		if err != nil {
			t.Errorf("RegisterClick failed on iteration %d: %v", i, err)
		}
	}

	// Verify the count in snapshot
	snapshot := repo.GetCountSnapshot()
	found := false
	for _, banner := range snapshot.Banners {
		if banner.BannerID == 1 {
			found = true
			if banner.Count != 6 { // 1 initial + 5 additional clicks
				t.Errorf("Expected count 6, got %d", banner.Count)
			}
			break
		}
	}
	if !found {
		t.Error("Banner not found in snapshot")
	}
}

func TestGetCountSnapshot(t *testing.T) {
	repo := setupTestRepository()

	// Register some clicks
	repo.RegisterClick(1)
	repo.RegisterClick(2)
	repo.RegisterClick(1)

	snapshot := repo.GetCountSnapshot()

	// Verify snapshot structure
	if snapshot.TimeStamp.IsZero() {
		t.Error("Snapshot timestamp is zero")
	}

	// Verify banner counts
	counts := make(map[int]int)
	for _, banner := range snapshot.Banners {
		counts[banner.BannerID] = banner.Count
	}

	if counts[1] != 2 {
		t.Errorf("Expected count 2 for banner 1, got %d", counts[1])
	}
	if counts[2] != 1 {
		t.Errorf("Expected count 1 for banner 2, got %d", counts[2])
	}
}

func TestZeroOutCounts(t *testing.T) {
	repo := setupTestRepository()

	// Register some clicks
	repo.RegisterClick(1)
	repo.RegisterClick(2)
	repo.RegisterClick(1)

	// Zero out counts
	repo.ZeroOutCounts()

	// Verify all counts are zero
	snapshot := repo.GetCountSnapshot()
	for _, banner := range snapshot.Banners {
		if banner.Count != 0 {
			t.Errorf("Expected count 0 for banner %d, got %d", banner.BannerID, banner.Count)
		}
	}
}

func TestGetValues(t *testing.T) {
	repo := setupTestRepository()

	// Register clicks for some banners
	repo.RegisterClick(1)
	repo.RegisterClick(2)
	repo.RegisterClick(1)

	// Get non-zero values
	values := repo.GetValues()

	// Verify we only get banners with non-zero counts
	if len(values) != 2 {
		t.Errorf("Expected 2 banners with non-zero counts, got %d", len(values))
	}

	// Verify the counts
	counts := make(map[int]int)
	for _, banner := range values {
		counts[banner.BannerID] = banner.Count
	}

	if counts[1] != 2 {
		t.Errorf("Expected count 2 for banner 1, got %d", counts[1])
	}
	if counts[2] != 1 {
		t.Errorf("Expected count 1 for banner 2, got %d", counts[2])
	}
}

func TestConcurrentClicks(t *testing.T) {
	repo := setupTestRepository()
	iterations := 100
	goroutines := 10

	// Create a channel to signal completion
	done := make(chan bool)

	// Launch multiple goroutines to register clicks concurrently
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				repo.RegisterClick(1)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Verify the final count
	snapshot := repo.GetCountSnapshot()
	found := false
	for _, banner := range snapshot.Banners {
		if banner.BannerID == 1 {
			found = true
			expectedCount := goroutines * iterations
			if banner.Count != expectedCount {
				t.Errorf("Expected count %d, got %d", expectedCount, banner.Count)
			}
			break
		}
	}
	if !found {
		t.Error("Banner not found in snapshot")
	}
}
