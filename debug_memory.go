package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"rsclabs-test/config"
	"rsclabs-test/internal/controller/http"
	"rsclabs-test/internal/model"
	"rsclabs-test/internal/repository"
	"rsclabs-test/internal/repository/inmemorystorage"
	"rsclabs-test/internal/service"
	"rsclabs-test/pkg/httpserver"
	"rsclabs-test/pkg/observe"
)

func TestDebugMemoryLeak(t *testing.T) {
	fmt.Println("=== DEBUGGING MEMORY LEAK ===")

	// 1. Test repository alone
	fmt.Println("\n1. Testing Repository Only:")
	testRepositoryMemory(t)

	// 2. Test service alone
	fmt.Println("\n2. Testing Service Only:")
	testServiceMemory(t)

	// 3. Test HTTP handlers
	fmt.Println("\n3. Testing HTTP Handlers:")
	testHTTPMemory(t)

	// 4. Test with fewer iterations
	fmt.Println("\n4. Testing Small Load:")
	testSmallLoad(t)
}

func testRepositoryMemory(t *testing.T) {
	l := observe.NewZapLogger("test")

	inMemoryStorage := inmemorystorage.NewInMemoryStorage(10, l)
	repo, err := repository.NewBannerRepository(inMemoryStorage)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 100; i++ {
		err := repo.RegisterClick(i % 10)
		if err != nil {
			t.Logf("Repository error: %v", err)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	growth := int64(m2.Alloc) - int64(m1.Alloc)
	fmt.Printf("   Repository memory growth: %d bytes (%d KB)\n", growth, growth/1024)

	if growth < 0 {
		fmt.Printf("   WARNING: Negative growth detected - possible integer overflow\n")
	}
}

func testServiceMemory(t *testing.T) {
	l := observe.NewZapLogger("test")
	cnf := config.NewConfig()
	server := httpserver.InitFiberServer(cnf.AppName)

	inMemoryStorage := inmemorystorage.NewInMemoryStorage(10, l)
	repo, err := repository.NewBannerRepository(inMemoryStorage)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	statsService := service.NewStatisticsService(repo, server, l)

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < 100; i++ {
		req := model.StatisticsRequest{
			BannerID: 1,
			From:     "2024-01-01",
			To:       "2024-01-31",
		}

		_, err := statsService.GetStatistics(context.Background(), req)
		if err != nil {
			l.Error(err)
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	growth := int64(m2.Alloc) - int64(m1.Alloc)
	fmt.Printf("   Service memory growth: %d bytes (%d KB)\n", growth, growth/1024)
}

func testHTTPMemory(t *testing.T) {
	app := setupTestApp()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Small HTTP test with proper cleanup
	for i := 0; i < 100; i++ {
		req1 := httptest.NewRequest("POST", "/click/1", nil)
		resp1, err := app.Test(req1)
		if err != nil {
			t.Logf("HTTP error: %v", err)
		} else {
			resp1.Body.Close() // IMPORTANT: Close response body
		}

		body := `{"from": "2024-01-01", "to": "2024-01-31"}`
		req2 := httptest.NewRequest("POST", "/stats/1", bytes.NewBufferString(body))
		req2.Header.Set("Content-Type", "application/json")
		resp2, err := app.Test(req2)
		if err != nil {
			t.Logf("HTTP error: %v", err)
		} else {
			resp2.Body.Close() // IMPORTANT: Close response body
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	growth := int64(m2.Alloc) - int64(m1.Alloc)
	fmt.Printf("   HTTP memory growth: %d bytes (%d KB)\n", growth, growth/1024)
}

func testSmallLoad(t *testing.T) {
	var memoryReadings []uint64

	app := setupSimpleApp()

	for iteration := 0; iteration < 5; iteration++ {
		var m runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m)
		memoryReadings = append(memoryReadings, m.Alloc)

		fmt.Printf("   Iteration %d: Memory = %d KB\n", iteration, m.Alloc/1024)

		for i := 0; i < 50; i++ {
			req := httptest.NewRequest("POST", "/click/1", nil)
			resp, err := app.Test(req)
			if err == nil {
				resp.Body.Close() // IMPORTANT: Close response body
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("   Memory progression:")
	for i, mem := range memoryReadings {
		fmt.Printf("     Reading %d: %d KB\n", i, mem/1024)
	}
}

func setupSimpleApp() *fiber.App {
	app := fiber.New()
	l := observe.NewZapLogger("test")

	inMemoryStorage := inmemorystorage.NewInMemoryStorage(10, l)
	repo, _ := repository.NewBannerRepository(inMemoryStorage)

	app.Post("/click/:bannerID", func(c *fiber.Ctx) error {
		repo.RegisterClick(0)
		return c.SendString("ok")
	})

	return app
}

func setupTestApp() *fiber.App {
	app := fiber.New()
	l := observe.NewZapLogger("test-app")
	cnf := config.NewConfig()
	server := httpserver.InitFiberServer(cnf.AppName)

	inMemoryStorage := inmemorystorage.NewInMemoryStorage(cnf.MaxBanners, l)

	bannerRepository, err := repository.NewBannerRepository(inMemoryStorage)
	if err != nil {
		l.Fatal("failed to create banner repository", map[string]any{"err": err})
	}

	statisticsService := service.NewStatisticsService(
		bannerRepository,
		server,
		l,
	)

	http.NewRouter(
		bannerRepository,
		statisticsService,
		server,
		l,
	)

	return app
}

func TestRepositoryImplementation(t *testing.T) {
	fmt.Println("=== TESTING REPOSITORY DETAILS ===")

	l := observe.NewZapLogger("test")
	inMemoryStorage := inmemorystorage.NewInMemoryStorage(5, l)
	repo, err := repository.NewBannerRepository(inMemoryStorage)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Check initial state
	snapshot := repo.GetCountSnapshot()
	fmt.Printf("Initial snapshot: %+v\n", snapshot)

	// Test multiple clicks
	for i := 0; i < 20; i++ {
		err := repo.RegisterClick(i % 5)
		fmt.Printf("Click %d (banner %d): error=%v\n", i, i%5, err)

		if i%5 == 0 {
			snapshot := repo.GetCountSnapshot()
			fmt.Printf("Snapshot at %d: %+v\n", i, snapshot)
		}
	}

	// Final snapshot
	finalSnapshot := repo.GetCountSnapshot()
	fmt.Printf("Final snapshot: %+v\n", finalSnapshot)
}

func TestIsolateProblem(t *testing.T) {
	fmt.Println("=== ISOLATING THE PROBLEM ===")

	// Test 1: Just memory reading
	fmt.Println("1. Testing memory reading accuracy:")
	for i := 0; i < 3; i++ {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("   Reading %d: Alloc=%d, TotalAlloc=%d\n", i, m.Alloc, m.TotalAlloc)
		time.Sleep(100 * time.Millisecond)
	}

	// Test 2: Check for overflow
	fmt.Println("2. Testing for integer overflow:")
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Allocate some memory intentionally
	data := make([]byte, 1024*1024) // 1MB
	_ = data

	runtime.ReadMemStats(&m2)

	diff := int64(m2.Alloc) - int64(m1.Alloc)
	unsignedDiff := m2.Alloc - m1.Alloc

	fmt.Printf("   Signed difference: %d\n", diff)
	fmt.Printf("   Unsigned difference: %d\n", unsignedDiff)
	fmt.Printf("   Expected ~1MB: %d\n", 1024*1024)

	if diff < 0 {
		fmt.Printf("   ERROR: Negative signed difference indicates overflow!\n")
	}
}

func TestMemoryLeakFixed(t *testing.T) {
	app := setupTestApp()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Run operations with proper cleanup
	iterations := 1000
	for i := 0; i < iterations; i++ {
		// Simulate HTTP requests with proper cleanup
		req := httptest.NewRequest("POST", "/click/1", nil)
		resp, err := app.Test(req)
		if err == nil {
			resp.Body.Close() // FIX: Close response body
		}

		// Also test stats endpoint
		statsBody := `{"from": "2024-01-01", "to": "2024-01-31"}`
		req2 := httptest.NewRequest("POST", "/stats/1", bytes.NewBufferString(statsBody))
		req2.Header.Set("Content-Type", "application/json")
		resp2, err := app.Test(req2)
		if err == nil {
			resp2.Body.Close() // FIX: Close response body
		}
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	allocDiff := int64(m2.Alloc) - int64(m1.Alloc)
	totalAllocDiff := int64(m2.TotalAlloc) - int64(m1.TotalAlloc)

	fmt.Printf("Memory before: %d KB\n", m1.Alloc/1024)
	fmt.Printf("Memory after: %d KB\n", m2.Alloc/1024)
	fmt.Printf("Memory difference: %d KB\n", allocDiff/1024)
	fmt.Printf("Total allocated difference: %d KB\n", totalAllocDiff/1024)

	// Check for overflow
	if allocDiff < 0 {
		t.Errorf("Integer overflow detected in memory calculation")
		return
	}

	maxAcceptableGrowth := int64(10 * 1024 * 1024) // 10MB
	if allocDiff > maxAcceptableGrowth {
		t.Errorf("Potential memory leak detected: %d bytes growth", allocDiff)
	}
}
