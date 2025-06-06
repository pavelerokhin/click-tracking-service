package inmemorystorage

import (
	"fmt"
	"sync"
	"time"

	"rsclabs-test/internal/model"
	"rsclabs-test/pkg/observe"
)

// InMemoryStorage is a thread-safe generic storage using an array
type InMemoryStorage struct {
	mux          sync.RWMutex
	maxCapacity  int
	values       map[int]*model.Banner
	minTimestamp time.Time
	maxTimestamp time.Time
	totalCount   int
	l            *observe.Logger
}

// NewInMemoryStorage creates a new instance of InMemoryStorage
func NewInMemoryStorage(maxCapacity int, l *observe.Logger) *InMemoryStorage {
	values := make(map[int]*model.Banner, maxCapacity)
	storage := InMemoryStorage{
		values:      values,
		maxCapacity: maxCapacity,
		l:           l,
	}

	storage.seedBanners()

	return &storage
}

func (s *InMemoryStorage) GetSnapshot() map[int]model.Banner {
	s.mux.RLock()
	defer s.mux.RUnlock()

	result := make(map[int]model.Banner)

	for _, banner := range s.values {
		if !banner.IsEmpty() {
			result[banner.BannerID] = model.Banner{
				TimeStamp: banner.TimeStamp,
				Name:      banner.Name,
				BannerID:  banner.BannerID,
				Count:     banner.Count,
			}
		}
	}

	return result
}

func (s *InMemoryStorage) Set(index int, value model.Banner) bool {
	s.mux.Lock()
	defer s.mux.Unlock()

	if index < 0 || index >= len(s.values) {
		return false
	}

	s.values[index] = &value

	return true
}

func (s *InMemoryStorage) IncrementCountTakeTimestamp(id int) error {
	s.mux.Lock()
	defer s.mux.Unlock()

	if id < 0 || id >= s.maxCapacity {
		return fmt.Errorf("invalid index: %d", id)
	}
	s.values[id].IncrementCount()

	t := time.Now()
	s.values[id].TimeStamp = t
	if s.totalCount == 0 {
		s.minTimestamp = t
	} else {
		s.maxTimestamp = t
	}

	s.totalCount += 1

	return nil
}

func (s *InMemoryStorage) ClearCount() {
	s.mux.Lock()
	defer s.mux.Unlock()

	s.values = make(map[int]*model.Banner, s.maxCapacity)
	s.seedBanners()
}

func (s *InMemoryStorage) GetNotZeroValues() []model.Banner {
	s.mux.RLock()
	defer s.mux.RUnlock()

	var result []model.Banner
	for _, v := range s.values {
		if !v.IsEmpty() {
			banner := model.Banner{
				TimeStamp: v.TimeStamp,
				Name:      v.Name,
				BannerID:  v.BannerID,
				Count:     v.Count,
			}
			result = append(result, banner)
		}
	}

	return result
}

func (s *InMemoryStorage) GetMaxCapacity() int {
	return s.maxCapacity
}
func (s *InMemoryStorage) seedBanners() {
	for i := 0; i < s.maxCapacity; i++ {
		s.values[i] = &model.Banner{
			TimeStamp: time.Now(),
			Name:      fmt.Sprintf("Banner %d", i),
			BannerID:  i,
			Count:     0,
		}
	}
}
