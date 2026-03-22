package jobs

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jaredwarren/ytdl/internal/download"
)

// Status of a download job.
type Status string

const (
	StatusQueued   Status = "queued"
	StatusRunning  Status = "running"
	StatusComplete Status = "completed"
	StatusFailed   Status = "failed"
)

// Job is a single YouTube download request.
type Job struct {
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	Status     Status    `json:"status"`
	Message    string    `json:"message,omitempty"`
	OutputPath string    `json:"output_path,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Store holds jobs and runs the worker.
type Store struct {
	mu     sync.RWMutex
	byID   map[string]*Job
	runner *download.Runner
	ch     chan string
}

// NewStore creates a job store and starts the background worker.
func NewStore(runner *download.Runner) *Store {
	s := &Store{
		byID:   make(map[string]*Job),
		runner: runner,
		ch:     make(chan string, 32),
	}
	go s.worker()
	return s
}

func (s *Store) worker() {
	for id := range s.ch {
		s.runOne(id)
	}
}

func (s *Store) runOne(id string) {
	s.mu.Lock()
	j, ok := s.byID[id]
	if !ok || j == nil {
		s.mu.Unlock()
		return
	}
	j.Status = StatusRunning
	j.UpdatedAt = time.Now()
	url := j.URL
	s.mu.Unlock()

	start := time.Now().Add(-2 * time.Second)
	ctx := context.Background()
	out, err := s.runner.Run(ctx, url, start)

	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok = s.byID[id]
	if !ok || j == nil {
		return
	}
	j.UpdatedAt = time.Now()
	if err != nil {
		j.Status = StatusFailed
		j.Message = err.Error()
		return
	}
	j.Status = StatusComplete
	j.OutputPath = out
}

// Create enqueues a new job. id must be unique.
func (s *Store) Create(id, rawURL string) (*Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.byID[id]; exists {
		return nil, errors.New("duplicate id")
	}
	now := time.Now()
	j := &Job{
		ID:        id,
		URL:       rawURL,
		Status:    StatusQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.byID[id] = j
	select {
	case s.ch <- id:
	default:
		delete(s.byID, id)
		return nil, errors.New("queue full")
	}
	return j, nil
}

// Get returns a copy of the job for JSON responses.
func (s *Store) Get(id string) (*Job, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.byID[id]
	if !ok {
		return nil, false
	}
	cp := *j
	return &cp, true
}
