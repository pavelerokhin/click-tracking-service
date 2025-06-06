package model

type StatisticsResponse struct {
	Stats []Banner `json:"stats"`
}

func (r *StatisticsResponse) IsEmpty() bool {
	return len(r.Stats) == 0
}
