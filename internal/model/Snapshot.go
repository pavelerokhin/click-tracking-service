package model

import "time"

type Snapshot struct {
	Banners   map[int]Banner // Map of BannerID to Banner
	TimeStamp time.Time
}

func (s *Snapshot) FilterByBannerID(bannerID int) (Banner, bool) {
	b, ok := s.Banners[bannerID]
	return b, ok
}

func (s *Snapshot) IsEmpty() bool {
	return len(s.Banners) == 0
}
