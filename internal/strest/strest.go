package strest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Configuration struct {
	URL         string
	Method      string
	Headers     []string
	Body        string
	Timeout     int
	Requests    int
	Concurrency int
	Warmup      int
}

type result struct {
	statusCount map[int]int64
	lat         latency
	errors      map[string]int64
}

type latency struct {
	min   time.Duration
	max   time.Duration
	count int64
	sum   time.Duration
}

type Strest struct {
	config     *Configuration
	result     *result
	workerChan chan struct{}
}

func NewStrest(config *Configuration) *Strest {
	return &Strest{
		config: config,
		result: &result{
			statusCount: make(map[int]int64),
			lat: latency{
				min:   0,
				max:   0,
				count: 0,
				sum:   0,
			},
			errors: make(map[string]int64),
		},
		workerChan: make(chan struct{}, config.Concurrency),
	}
}

func (s *Strest) Run() (*Result, error) {
	client := http.Client{
		Timeout: time.Duration(s.config.Timeout) * time.Second,
	}
	fmt.Println("Warming up...")
	for range s.config.Warmup {
		req, err := s.getRequest()
		if err != nil {
			continue
		}
		client.Do(req)
	}
	fmt.Println("Warmup complete")

	var inChan = make(chan *http.Request, s.config.Concurrency)
	var outChan = make(chan *RequestResult, s.config.Requests)
	var totalTime time.Duration
	go func() {
		start := time.Now()
		wg := sync.WaitGroup{}
		for req := range inChan {
			s.workerChan <- struct{}{}
			wg.Add(1)
			go s.work(client, req, outChan, &wg)
		}
		wg.Wait()
		totalTime = time.Since(start)
		close(outChan)
	}()
	fmt.Println("Starting load test...")
	for range s.config.Requests {
		req, err := s.getRequest()
		if err != nil {
			continue
		}
		inChan <- req
	}
	close(inChan)
	s.processOut(outChan)
	fmt.Println("Finished")
	return &Result{
		Min:       s.result.lat.min,
		Max:       s.result.lat.max,
		Average:   s.result.lat.average(),
		Statuses:  s.result.statusCount,
		Errors:    s.result.errors,
		TotalTime: totalTime,
	}, nil
}

func (s *Strest) getRequest() (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	if s.config.Body != "" {
		body, err := json.Marshal(s.config.Body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(s.config.Method, s.config.URL, bytes.NewBuffer(body))
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest(s.config.Method, s.config.URL, nil)
		if err != nil {
			return nil, err
		}
	}
	for _, h := range s.config.Headers {
		header := strings.Split(h, ":")
		req.Header.Set(header[0], header[1])
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

type Result struct {
	Min       time.Duration
	Max       time.Duration
	Average   time.Duration
	Statuses  map[int]int64
	Errors    map[string]int64
	TotalTime time.Duration
}

func (r *Result) String() string {
	var totalErrors int64 = 0
	for _, v := range r.Errors {
		totalErrors += v
	}
	var totalSucesses int64 = 0
	for _, v := range r.Statuses {
		totalSucesses += v
	}
	totalRequests := totalSucesses + totalErrors
	header := fmt.Sprintf("\nResults:\n\n\tTotal requests: %d. Of which %d errored and %d went through.\n\tTest execution time: %v", totalRequests, totalErrors, totalSucesses, r.TotalTime)

	latency := fmt.Sprintf("\n\nLatency:\n\tMin: %v, Max: %v, Average: %v", r.Min, r.Max, r.Average)
	var statuses string = "\n\nStatus:"
	for k, v := range r.Statuses {
		statuses += fmt.Sprintf("\n\t%d: %d", k, v)
	}
	var errors string = "\n\nErrors:"
	for k, v := range r.Errors {
		errors += fmt.Sprintf("\n\t%s: %d", k, v)
	}
	return header + statuses + latency + errors
}

type RequestResult struct {
	StatusCode int
	Error      error
	Elapsed    time.Duration
}

func (s *Strest) processOut(outChan <-chan *RequestResult) {
	for res := range outChan {
		if res.Error != nil {
			s.result.errors[res.Error.Error()]++
			continue
		}
		s.result.lat.add(res.Elapsed)
		s.result.statusCount[res.StatusCode]++
	}
}

func (s *Strest) work(client http.Client, req *http.Request, outChan chan<- *RequestResult, wg *sync.WaitGroup) {
	start := time.Now()
	res, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		outChan <- &RequestResult{
			Error:   err,
			Elapsed: elapsed,
		}
	} else {
		outChan <- &RequestResult{
			StatusCode: res.StatusCode,
			Error:      nil,
			Elapsed:    elapsed,
		}
	}
	wg.Done()
	<-s.workerChan
}

func (l *latency) add(elapsed time.Duration) {
	l.count++
	l.sum += elapsed
	if elapsed < l.min || l.min == 0 {
		l.min = elapsed
	}
	if elapsed > l.max {
		l.max = elapsed
	}
}

func (l *latency) average() time.Duration {
	if l.count == 0 {
		return 0
	}
	return time.Duration(l.sum.Nanoseconds() / l.count)
}
