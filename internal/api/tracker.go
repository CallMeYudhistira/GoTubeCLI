package api

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

type JobStatus string

const (
	StatusPending     JobStatus = "pending"
	StatusDownloading JobStatus = "downloading"
	StatusMerging     JobStatus = "merging"
	StatusCompleted   JobStatus = "completed"
	StatusError       JobStatus = "error"
)

type JobInfo struct {
	ID       string    `json:"id"`
	Progress int       `json:"progress"`
	Status   JobStatus `json:"status"`
	File     string    `json:"file,omitempty"`
	Error    string    `json:"error,omitempty"`
}

type JobTracker struct {
	mu   sync.RWMutex
	jobs map[string]*JobInfo
}

var Tracker = &JobTracker{
	jobs: make(map[string]*JobInfo),
}

func GenerateJobID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (t *JobTracker) CreateJob() *JobInfo {
	id := GenerateJobID()
	t.mu.Lock()
	defer t.mu.Unlock()

	job := &JobInfo{
		ID:       id,
		Progress: 0,
		Status:   StatusPending,
	}
	t.jobs[id] = job
	return job
}

func (t *JobTracker) GetJob(id string) (*JobInfo, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	job, exists := t.jobs[id]
	return job, exists
}

func (t *JobTracker) UpdateProgress(id string, progress int, status JobStatus) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if job, exists := t.jobs[id]; exists {
		job.Progress = progress
		if status != "" {
			job.Status = status
		}
	}
}

func (t *JobTracker) CompleteJob(id string, file string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if job, exists := t.jobs[id]; exists {
		job.Progress = 100
		job.Status = StatusCompleted
		job.File = file
	}
}

func (t *JobTracker) FailJob(id string, err string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if job, exists := t.jobs[id]; exists {
		job.Status = StatusError
		job.Error = err
	}
}

// CustomWriter implements io.Writer to report progress back to the tracker
type CustomWriter struct {
	JobID       string
	TotalSize   int64
	CurrentSize int64
}

func (cw *CustomWriter) SetTotal(size int64) {
	cw.TotalSize = size
}

func (cw *CustomWriter) SetProgress(percent float64) {
	progress := int(percent)
	if progress > 100 {
		progress = 100
	}
	Tracker.UpdateProgress(cw.JobID, progress, StatusDownloading)
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	n = len(p)
	cw.CurrentSize += int64(n)
	if cw.TotalSize > 0 {
		progress := int(float64(cw.CurrentSize) / float64(cw.TotalSize) * 100)
		Tracker.UpdateProgress(cw.JobID, progress, StatusDownloading)
	}
	return n, nil
}

