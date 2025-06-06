# Banner Click Counter Service

A high-performance Go service for tracking banner clicks with per-minute statistics aggregation.

## Features

- Track banner clicks in real-time
- Get statistics for specific banners
- In-memory storage with periodic snapshots
- Memory leak detection and monitoring

## Overview

This service implements a banner click counting system that:
- Tracks clicks for banners with IDs 1-100
- Aggregates click data into per-minute statistics
- Handles high-frequency requests (500+ RPS)
- Uses thread-safe in-memory storage

## Architecture

```
HTTP Handler → Service Layer → Storage Layer
    ↓              ↓              ↓
Counter API    Statistics     In-Memory
Stats API     Click Counter   Thread-Safe
```

## Performance Characteristics

**Load Testing Results:**
- **Junior Level (10-50 RPS)**: PASSED - 200 requests at 20 RPS
- **Middle+ Level (100-500 RPS)**: PASSED - 500 requests at 100 RPS
- **Success Rate**: 100% under all tested loads
- **Concurrency**: Thread-safe concurrent request handling
- **Response Time**: <10ms average for counter requests

## API Endpoints

### 1. Increment Counter
`GET /counter/{bannerID}`

Increments click counter for specified banner (ID: 1-100).

**Example:**
```bash
curl -X GET http://localhost:8080/counter/12
```

**Response:**
```json
{"bannerID": 12, "success": true}
```

### 2. Get Statistics
`POST /stats/{bannerID}`

Retrieves per-minute click statistics for specified time range.

**Request:**
```json
{
  "from": "2025-06-06T01:00:00",
  "to": "2025-06-06T02:00:00"
}
```

**Example:**
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"from": "2025-06-06T01:00:00", "to": "2025-06-06T02:00:00"}' \
  http://localhost:8080/stats/12
```

**Response:**
```json
{
  "stats": [
    {
      "ts": "2025-06-06T01:30:00.123456+02:00",
      "name": "Banner 12",
      "v": 15
    }
  ]
}
```

## Installation

1. **Build:**
```bash
go mod tidy
go build -o banner-counter ./cmd/app
```

2. **Run:**
```bash
./banner-counter
```

Service starts on `http://localhost:8080`

## Load Testing

### Basic Test (20 RPS)
```bash
for i in {1..200}; do
  curl -X GET http://localhost:8080/counter/12 &
  if [ $((i % 20)) -eq 0 ]; then sleep 1; fi
done
wait
```

### High Load Test (100 RPS)
```bash
for i in {1..500}; do
  curl -X GET http://localhost:8080/counter/12 &
  if [ $((i % 100)) -eq 0 ]; then sleep 1; fi
done
wait
```

### Verify Statistics
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"from": "2025-06-06T00:00:00", "to": "2025-06-06T23:59:59"}' \
  http://localhost:8080/stats/12
```

## Configuration

Environment variables:
- `PORT`: Server port (default: 8080)
- `MAX_BANNERS`: Maximum banner count (default: 100)
- `LOG_LEVEL`: debug, info, warn, error

## Error Handling

**Invalid Banner ID:**
```bash
curl -X GET http://localhost:8080/counter/101
# Returns: 400 Bad Request
```

**Invalid Time Format:**
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"from": "invalid-date"}' http://localhost:8080/stats/12
# Returns: 400 Bad Request
```

**No Data Found:**
```bash
# Returns: {"stats": []}
```

## Implementation Details

**Data Structures:**
- Banner IDs: 1-100 (API) → 0-99 (internal array)
- Thread safety: Mutex-protected operations
- Time handling: Local input → UTC storage
- Aggregation: Per-minute click grouping

**Key Features:**
- Concurrent request handling
- Timezone-aware time parsing
- Input validation and sanitization
- Graceful error responses

## Monitoring

**Health Check:**
```bash
curl -X GET http://localhost:8080/counter/1
```

**Performance Monitoring:**
```bash
# Monitor during load testing
top -p $(pgrep banner-counter)
netstat -an | grep :8080 | wc -l
```

## Testing Results Summary 

| Test Type      | Target | Result | Status |
|----------------|--------|--------|--------|
| level 1        | 10-50 RPS | 20 RPS (200 req) | PASSED |
| level 2        | 100-500 RPS | 100 RPS (500 req) | PASSED |
| Concurrency    | Multiple simultaneous | Thread-safe | PASSED |
| Data Accuracy  | Click counting | 100% accurate | PASSED |
| Error Handling | Invalid inputs | Proper responses | PASSED |

## Memory Leak Testing

The service includes comprehensive memory leak testing to ensure stable long-term operation. The test suite includes:

- `TestMemoryLeaks`: Basic memory leak detection
- `TestGoroutineLeaks`: Goroutine leak detection
- `TestStressWithMemoryMonitoring`: Stress testing with memory monitoring
- `TestLongRunningMemoryLeak`: Extended duration memory leak testing

## Race Condition Testing

The service includes race condition detection to ensure thread safety. To run the tests with race detection:

```bash
go test -race ./...
```

The test suite includes:
- `TestRaceConditions`: Simulates concurrent access to shared resources
- Race-aware request handling in all endpoints
- Proper synchronization for concurrent operations

Key features of race condition testing:
- Multiple concurrent goroutines (10 goroutines)
- High-frequency concurrent requests (100 requests per goroutine)
- Proper resource cleanup and synchronization
- Channel-based coordination between goroutines

### Memory Test Parameters

- Test duration: 60 seconds
- Measurement interval: 10 seconds
- Maximum acceptable memory growth: 100MB
- Growth rate threshold: 80%

### Running Memory Tests

To run all memory tests:
```bash
go test -v ./...
```

To skip long-running tests:
```bash
go test -v -short ./...
```

### Memory Test Output

The memory tests provide detailed output including:
- Memory allocation (KB)
- System memory usage (KB)
- Number of garbage collections
- Memory growth rate
- Total memory growth

## License

MIT

