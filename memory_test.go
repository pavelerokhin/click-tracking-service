package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	_ "net/http/pprof" // Import pprof for profiling
	"runtime"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Memory leak test using runtime.MemStats
func TestMemoryLeaks(t *testing.T) {
	// Setup your application
	app := setupTestApp()

	// Get initial memory stats
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Run operations that might leak memory
	runLoadTest(app, 1000)

	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocDiff := m2.Alloc - m1.Alloc
	totalAllocDiff := m2.TotalAlloc - m1.TotalAlloc

	fmt.Printf("Memory before: %d KB\n", m1.Alloc/1024)
	fmt.Printf("Memory after: %d KB\n", m2.Alloc/1024)
	fmt.Printf("Memory difference: %d KB\n", allocDiff/1024)
	fmt.Printf("Total allocated difference: %d KB\n", totalAllocDiff/1024)

	maxAcceptableGrowth := uint64(10 * 1024 * 1024) // 10MB
	if allocDiff > maxAcceptableGrowth {
		t.Errorf("Potential memory leak detected: %d bytes growth", allocDiff)
	}
}

// Goroutine leak test
func TestGoroutineLeaks(t *testing.T) {
	initialGoroutines := runtime.NumGoroutine()

	// Run operations that might spawn goroutines
	app := setupTestApp()
	runLoadTest(app, 100)

	// Wait for goroutines to finish
	time.Sleep(2 * time.Second)
	runtime.GC()

	finalGoroutines := runtime.NumGoroutine()
	goroutineDiff := finalGoroutines - initialGoroutines

	fmt.Printf("Goroutines before: %d\n", initialGoroutines)
	fmt.Printf("Goroutines after: %d\n", finalGoroutines)
	fmt.Printf("Goroutine difference: %d\n", goroutineDiff)

	// Allow some tolerance for background goroutines
	maxAcceptableGoroutines := 5
	if goroutineDiff > maxAcceptableGoroutines {
		t.Errorf("Potential goroutine leak: %d extra goroutines", goroutineDiff)
	}
}

// Stress test with memory monitoring
func TestStressWithMemoryMonitoring(t *testing.T) {
	app := setupTestApp()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Monitor memory usage during stress test
	go monitorMemoryUsage(ctx, t)

	// Run stress test
	for i := 0; i < 10000; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			runSingleRequest(app)

			// Force GC every 1000 iterations
			if i%1000 == 0 {
				runtime.GC()
			}
		}
	}
}

// Long-running memory leak test
func TestLongRunningMemoryLeak(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running test in short mode")
	}

	app := setupTestApp()

	// Record memory stats over time
	var memoryStats []uint64
	duration := 60 * time.Second
	interval := 10 * time.Second // Increased interval to allow for GC

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	timeout := time.After(duration)

	// Force initial GC
	runtime.GC()

	for {
		select {
		case <-ticker.C:
			// Run some operations
			runLoadTest(app, 50)

			// Force GC before measuring
			runtime.GC()

			// Record memory usage
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			memoryStats = append(memoryStats, m.Alloc)

			// Print detailed memory stats
			fmt.Printf("Memory stats - Alloc: %d KB, Sys: %d KB, NumGC: %d\n",
				m.Alloc/1024, m.Sys/1024, m.NumGC)

		case <-timeout:
			// Analyze memory trend
			analyzeMemoryTrend(t, memoryStats)
			return
		}
	}
}

func runLoadTest(app *fiber.App, iterations int) {
	for i := 0; i < iterations; i++ {
		runSingleRequest(app)
	}
}

func runSingleRequest(app *fiber.App) {
	// Simulate HTTP requests with proper synchronization
	req := httptest.NewRequest("POST", "/click/1", nil)
	resp, err := app.Test(req)
	if err == nil {
		resp.Body.Close() // Close response body
	}

	// Also test stats endpoint with proper synchronization
	statsBody := `{"from": "2024-01-01", "to": "2024-01-31"}`
	req2 := httptest.NewRequest("POST", "/stats/1",
		bytes.NewBufferString(statsBody))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2)
	if err == nil {
		resp2.Body.Close() // Close response body
	}
}

func monitorMemoryUsage(ctx context.Context, t *testing.T) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	var maxMemory uint64

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Peak memory usage: %d KB\n", maxMemory/1024)
			return
		case <-ticker.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)

			if m.Alloc > maxMemory {
				maxMemory = m.Alloc
			}

			// Alert if memory usage is too high
			if m.Alloc > 100*1024*1024 { // 100MB
				t.Errorf("High memory usage detected: %d KB", m.Alloc/1024)
			}
		}
	}
}

func analyzeMemoryTrend(t *testing.T, stats []uint64) {
	if len(stats) < 3 {
		t.Log("Not enough data points for trend analysis")
		return
	}

	// Print all measurements for debugging
	fmt.Println("\nMemory measurements (KB):")
	for i, stat := range stats {
		fmt.Printf("Measurement %d: %d\n", i, stat/1024)
	}

	// Simple trend analysis: check if memory is consistently growing
	growthCount := 0
	for i := 1; i < len(stats); i++ {
		if stats[i] > stats[i-1] {
			growthCount++
		}
	}

	growthRate := float64(growthCount) / float64(len(stats)-1)

	fmt.Printf("\nMemory growth rate: %.2f%%\n", growthRate*100)

	// If memory grows in more than 80% of measurements, flag as potential leak
	// Increased threshold to be more realistic
	if growthRate > 0.8 {
		t.Errorf("Potential memory leak: memory growing in %.2f%% of measurements", growthRate*100)
	}

	// Check absolute growth using signed arithmetic to avoid overflow
	totalGrowth := int64(stats[len(stats)-1]) - int64(stats[0])
	fmt.Printf("Total memory growth: %d KB\n", totalGrowth/1024)

	// Only fail if we see significant growth
	if totalGrowth > 100*1024*1024 { // 100MB
		t.Errorf("Excessive memory growth: %d KB", totalGrowth/1024)
	}
}

// Benchmark test with memory allocation tracking
func BenchmarkMemoryAllocation(b *testing.B) {
	app := setupTestApp()

	b.ResetTimer()
	b.ReportAllocs() // Report memory allocations

	for i := 0; i < b.N; i++ {
		runSingleRequest(app)
	}
}

// TestRaceConditions checks for potential race conditions in the application
func TestRaceConditions(t *testing.T) {
	app := setupTestApp()

	// Create a channel to coordinate goroutines
	done := make(chan bool)

	// Launch multiple goroutines that will access shared resources
	for i := 0; i < 10; i++ {
		go func() {
			// Simulate concurrent requests
			for j := 0; j < 100; j++ {
				runSingleRequest(app)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
