package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type RequestConfig struct {
	URL     string
	Method  string
	Headers map[string]string
	Body    string
}

type Result struct {
	Duration time.Duration
	Status   int
	Error    error
}

type Stats struct {
	TotalRequests int
	SuccessCount  int
	ErrorCount    int
	TotalDuration time.Duration
	MinDuration   time.Duration
	MaxDuration   time.Duration
	StatusCodes   map[int]int
	mutex         sync.Mutex
}

func NewStats() *Stats {
	return &Stats{
		StatusCodes: make(map[int]int),
		MinDuration: time.Hour, // Initialize with large value
	}
}

func (s *Stats) AddResult(result Result) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.TotalRequests++
	s.TotalDuration += result.Duration

	if result.Error != nil {
		s.ErrorCount++
	} else {
		s.SuccessCount++
		s.StatusCodes[result.Status]++

		if result.Duration < s.MinDuration {
			s.MinDuration = result.Duration
		}
		if result.Duration > s.MaxDuration {
			s.MaxDuration = result.Duration
		}
	}
}

func (s *Stats) Print() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	fmt.Println("\n=== Load Test Results ===")
	fmt.Printf("Total Requests: %d\n", s.TotalRequests)
	fmt.Printf("Successful: %d\n", s.SuccessCount)
	fmt.Printf("Failed: %d\n", s.ErrorCount)
	fmt.Printf("Average Duration: %v\n", s.TotalDuration/time.Duration(s.TotalRequests))
	fmt.Printf("Min Duration: %v\n", s.MinDuration)
	fmt.Printf("Max Duration: %v\n", s.MaxDuration)
	fmt.Println("Status Codes:")
	for code, count := range s.StatusCodes {
		fmt.Printf("  %d: %d\n", code, count)
	}
}

func makeRequest(config RequestConfig) Result {
	start := time.Now()
	var result Result

	// Create request body if provided
	var reqBody *bytes.Buffer
	if config.Body != "" {
		reqBody = bytes.NewBufferString(config.Body)
	} else {
		reqBody = bytes.NewBufferString("")
	}

	req, err := http.NewRequest(config.Method, config.URL, reqBody)
	if err != nil {
		result.Error = err
		return result
	}

	// Set headers
	for key, value := range config.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
		return result
	}
	defer resp.Body.Close()

	// Read response body (optional)
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		return result
	}

	result.Duration = time.Since(start)
	result.Status = resp.StatusCode
	return result
}

func worker(id int, config RequestConfig, requests chan struct{}, results chan Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for range requests {
		results <- makeRequest(config)
	}
}

func loadTest(config RequestConfig, concurrency int, totalRequests int) {
	requests := make(chan struct{}, totalRequests)
	results := make(chan Result, totalRequests)
	stats := NewStats()
	var wg sync.WaitGroup

	// Create worker pool
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go worker(i, config, requests, results, &wg)
	}

	// Send requests
	for i := 0; i < totalRequests; i++ {
		requests <- struct{}{}
	}
	close(requests)

	// Collect results
	go func() {
		for result := range results {
			stats.AddResult(result)
		}
	}()

	wg.Wait()
	close(results)

	stats.Print()
}

func main() {
	url := flag.String("url", "", "Target URL to test")
	method := flag.String("method", "GET", "HTTP method")
	concurrency := flag.Int("concurrency", 10, "Number of concurrent workers")
	requests := flag.Int("requests", 100, "Total number of requests")
	body := flag.String("body", "", "Request body (JSON)")
	header := flag.String("header", "", "Custom headers (JSON)")
	flag.Parse()

	if *url == "" {
		log.Fatal("URL is required")
	}

	config := RequestConfig{
		URL:    *url,
		Method: *method,
	}

	// Parse headers if provided
	if *header != "" {
		var headers map[string]string
		if err := json.Unmarshal([]byte(*header), &headers); err != nil {
			log.Fatalf("Invalid header JSON: %v", err)
		}
		config.Headers = headers
	}

	// Set body if provided
	if *body != "" {
		config.Body = *body
	}

	fmt.Printf("Starting load test:\n")
	fmt.Printf("  URL: %s\n", config.URL)
	fmt.Printf("  Method: %s\n", config.Method)
	fmt.Printf("  Concurrency: %d\n", *concurrency)
	fmt.Printf("  Total Requests: %d\n", *requests)
	if len(config.Headers) > 0 {
		fmt.Println("  Headers:", config.Headers)
	}
	if config.Body != "" {
		fmt.Println("  Body:", config.Body)
	}

	loadTest(config, *concurrency, *requests)
}
