package model

type StatisticsRequest struct {
	From     string `json:"from"`
	To       string `json:"to"`
	BannerID int    `json:"banner_id"`
}
