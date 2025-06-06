package model

import "time"

type Banner struct {
	TimeStamp time.Time `json:"ts"`
	Name      string    `json:"name"`
	BannerID  int       `json:"-"`
	Count     int       `json:"v"`
}

func NewBanner(bannerID int, name string) Banner {
	return Banner{
		BannerID:  bannerID,
		Name:      name,
		Count:     0,
		TimeStamp: time.Now(),
	}
}

func (b *Banner) IncrementCount() {
	b.Count++
}

func (b *Banner) IsEmpty() bool {
	return b.Count == 0
}
