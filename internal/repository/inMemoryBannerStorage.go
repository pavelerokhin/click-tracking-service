package repository

import (
	"rsclabs-test/internal/model"
	"rsclabs-test/internal/repository/inmemorystorage"
	"time"
)

type BannerRepositoryInMemory struct {
	storage    *inmemorystorage.InMemoryStorage
	MaxBanners int
}

func NewBannerRepository(storage *inmemorystorage.InMemoryStorage) (*BannerRepositoryInMemory, error) {
	return &BannerRepositoryInMemory{
		storage:    storage,
		MaxBanners: storage.GetMaxCapacity(),
	}, nil
}

func (r *BannerRepositoryInMemory) RegisterClick(id int) error {
	return r.storage.IncrementCountTakeTimestamp(id)
}

func (r *BannerRepositoryInMemory) GetCountSnapshot() model.Snapshot {
	return model.Snapshot{
		Banners:   r.storage.GetSnapshot(),
		TimeStamp: time.Now(),
	}
}

func (r *BannerRepositoryInMemory) ZeroOutCounts() {
	r.storage.ClearCount()
}

func (r *BannerRepositoryInMemory) GetValues() []model.Banner {
	return r.storage.GetNotZeroValues()
}
