package service

//
//import (
//	"reflect"
//	"testing"
//	"time"
//
//	"rsclabs-test/internal/model"
//)
//
//func TestFindTheEarliestTimestamp(t *testing.T) {
//	tests := []struct {
//		name      string
//		snapshots []model.Snapshot
//		expected  time.Time
//	}{
//		{
//			name:      "empty snapshots slice",
//			snapshots: []model.Snapshot{},
//			expected:  time.Time{},
//		},
//		{
//			name: "single snapshot with banners",
//			snapshots: []model.Snapshot{
//				{
//					Banners: []model.Banner{
//						{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//						{BannerID: 2, TimeStamp: time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC)},
//						{BannerID: 3, TimeStamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)},
//					},
//				},
//			},
//			expected: time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
//		},
//		{
//			name: "single snapshot with empty banners",
//			snapshots: []model.Snapshot{
//				{
//					Banners: []model.Banner{},
//				},
//			},
//			expected: time.Time{},
//		},
//		{
//			name: "multiple snapshots - only first is used",
//			snapshots: []model.Snapshot{
//				{
//					Banners: []model.Banner{
//						{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//					},
//				},
//				{
//					Banners: []model.Banner{
//						{BannerID: 2, TimeStamp: time.Date(2022, 1, 1, 8, 0, 0, 0, time.UTC)},
//					},
//				},
//			},
//			expected: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result := FindTheEarliestTimestamp(tt.snapshots)
//			if !result.Equal(tt.expected) {
//				t.Errorf("FindTheEarliestTimestamp() = %v, want %v", result, tt.expected)
//			}
//		})
//	}
//}
//
//func TestFindFirstTimestampInSnapshot(t *testing.T) {
//	// Mock current time for predictable testing
//	mockTime := time.Date(2023, 6, 1, 12, 0, 0, 0, time.UTC)
//
//	tests := []struct {
//		name     string
//		snapshot []model.Banner
//		expected time.Time
//	}{
//		{
//			name:     "empty snapshot",
//			snapshot: []model.Banner{},
//			expected: time.Time{},
//		},
//		{
//			name: "single non-empty banner",
//			snapshot: []model.Banner{
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//			},
//			expected: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
//		},
//		{
//			name: "multiple non-empty banners",
//			snapshot: []model.Banner{
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 2, TimeStamp: time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC)},
//				{BannerID: 3, TimeStamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)},
//			},
//			expected: time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
//		},
//		{
//			name: "mix of empty and non-empty banners",
//			snapshot: []model.Banner{
//				{}, // empty
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 2, TimeStamp: time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC)},
//			},
//			expected: time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
//		},
//		{
//			name: "banners with future timestamps",
//			snapshot: []model.Banner{
//				{BannerID: 1, TimeStamp: mockTime.Add(time.Hour)},                     // future
//				{BannerID: 2, TimeStamp: time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC)}, // past
//			},
//			expected: time.Date(2023, 1, 1, 8, 0, 0, 0, time.UTC),
//		},
//		{
//			name: "banners with same timestamp",
//			snapshot: []model.Banner{
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 2, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//			},
//			expected: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result := FindFirstTimestampInSnapshot(tt.snapshot)
//			if !result.Equal(tt.expected) {
//				t.Errorf("FindFirstTimestampInSnapshot() = %v, want %v", result, tt.expected)
//			}
//		})
//	}
//}
//
//func TestFilterByBannerIDInSnapshot(t *testing.T) {
//	tests := []struct {
//		name     string
//		snapshot []model.Banner
//		bannerID int
//		expected []model.Banner
//	}{
//		{
//			name:     "empty snapshot",
//			snapshot: []model.Banner{},
//			bannerID: 1,
//			expected: nil,
//		},
//		{
//			name: "no matching banners",
//			snapshot: []model.Banner{
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 2, TimeStamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)},
//			},
//			bannerID: 3,
//			expected: nil,
//		},
//		{
//			name: "single matching banner",
//			snapshot: []model.Banner{
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 2, TimeStamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)},
//			},
//			bannerID: 1,
//			expected: []model.Banner{
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//			},
//		},
//		{
//			name: "multiple matching banners",
//			snapshot: []model.Banner{
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 2, TimeStamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)},
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)},
//				{BannerID: 3, TimeStamp: time.Date(2023, 1, 1, 13, 0, 0, 0, time.UTC)},
//			},
//			bannerID: 1,
//			expected: []model.Banner{
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)},
//			},
//		},
//		{
//			name: "all banners match",
//			snapshot: []model.Banner{
//				{BannerID: 5, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 5, TimeStamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)},
//			},
//			bannerID: 5,
//			expected: []model.Banner{
//				{BannerID: 5, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 5, TimeStamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)},
//			},
//		},
//		{
//			name: "filter by zero ID",
//			snapshot: []model.Banner{
//				{BannerID: 0, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//				{BannerID: 1, TimeStamp: time.Date(2023, 1, 1, 11, 0, 0, 0, time.UTC)},
//			},
//			bannerID: 0,
//			expected: []model.Banner{
//				{BannerID: 0, TimeStamp: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)},
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			result := FilterByBannerIDInSnapshot(tt.snapshot, tt.bannerID)
//			if !reflect.DeepEqual(result, tt.expected) {
//				t.Errorf("FilterByBannerIDInSnapshot() = %v, want %v", result, tt.expected)
//			}
//		})
//	}
//}
//
//// Benchmark tests
//func BenchmarkFindFirstTimestampInSnapshot(b *testing.B) {
//	// Create a large snapshot for benchmarking
//	snapshot := make([]model.Banner, 1000)
//	for i := 0; i < 1000; i++ {
//		snapshot[i] = model.Banner{
//			BannerID:  i + 1,
//			TimeStamp: time.Date(2023, 1, 1, 0, 0, i, 0, time.UTC),
//		}
//	}
//
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		FindFirstTimestampInSnapshot(snapshot)
//	}
//}
//
//func BenchmarkFilterByBannerIDInSnapshot(b *testing.B) {
//	// Create a large snapshot for benchmarking
//	snapshot := make([]model.Banner, 1000)
//	for i := 0; i < 1000; i++ {
//		snapshot[i] = model.Banner{
//			BannerID:  i%10 + 1, // Creates repeated banner IDs
//			TimeStamp: time.Date(2023, 1, 1, 0, 0, i, 0, time.UTC),
//		}
//	}
//
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		FilterByBannerIDInSnapshot(snapshot, 5)
//	}
//}
