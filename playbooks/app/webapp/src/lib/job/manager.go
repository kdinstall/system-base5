package job

import (
	"context"
	"sync"
	"time"
)

// Job represents an installation job
type Job struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Status     string             `json:"status"` // pending, running, completed, failed
	StartTime  time.Time          `json:"start_time"`
	EndTime    time.Time          `json:"end_time"`
	Logs       []string           `json:"logs"`
	Error      string             `json:"error"`
	CancelFunc context.CancelFunc `json:"-"`
	mu         sync.Mutex
}

// JobManager manages all installation jobs
type JobManager struct {
	jobs map[string]*Job
	mu   sync.RWMutex
}

var (
	manager *JobManager
	once    sync.Once
)

// GetManager returns the singleton JobManager instance
func GetManager() *JobManager {
	once.Do(func() {
		manager = &JobManager{
			jobs: make(map[string]*Job),
		}
	})
	return manager
}

// CreateJob creates a new job
func (jm *JobManager) CreateJob(id, name string) *Job {
	jm.mu.Lock()
	defer jm.mu.Unlock()

	job := &Job{
		ID:        id,
		Name:      name,
		Status:    "pending",
		StartTime: time.Now(),
		Logs:      []string{},
	}
	jm.jobs[id] = job
	return job
}

// GetJob retrieves a job by ID
func (jm *JobManager) GetJob(id string) *Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return jm.jobs[id]
}

// ListJobs returns all jobs
func (jm *JobManager) ListJobs() []*Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	jobs := make([]*Job, 0, len(jm.jobs))
	for _, job := range jm.jobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// GetRunningJob returns the currently running job, or nil if none
func (jm *JobManager) GetRunningJob() *Job {
	jm.mu.RLock()
	defer jm.mu.RUnlock()

	for _, job := range jm.jobs {
		if job.Status == "running" || job.Status == "pending" {
			return job
		}
	}
	return nil
}

// AppendLog adds a log line to a job
func (jm *JobManager) AppendLog(id, line string) {
	jm.mu.RLock()
	job := jm.jobs[id]
	jm.mu.RUnlock()

	if job != nil {
		job.mu.Lock()
		job.Logs = append(job.Logs, line)
		job.mu.Unlock()
	}
}

// UpdateStatus updates a job's status
func (jm *JobManager) UpdateStatus(id, status string) {
	jm.mu.RLock()
	job := jm.jobs[id]
	jm.mu.RUnlock()

	if job != nil {
		job.mu.Lock()
		job.Status = status
		if status == "completed" || status == "failed" {
			job.EndTime = time.Now()
		}
		job.mu.Unlock()
	}
}

// SetError sets the error message for a job
func (jm *JobManager) SetError(id, errMsg string) {
	jm.mu.RLock()
	job := jm.jobs[id]
	jm.mu.RUnlock()

	if job != nil {
		job.mu.Lock()
		job.Error = errMsg
		job.mu.Unlock()
	}
}

// SetCancelFunc sets the cancel function for a job
func (jm *JobManager) SetCancelFunc(id string, cancelFunc context.CancelFunc) {
	jm.mu.RLock()
	job := jm.jobs[id]
	jm.mu.RUnlock()

	if job != nil {
		job.mu.Lock()
		job.CancelFunc = cancelFunc
		job.mu.Unlock()
	}
}

// GetLogs returns logs for a job starting from an offset
func (jm *JobManager) GetLogs(id string, offset int) []string {
	jm.mu.RLock()
	job := jm.jobs[id]
	jm.mu.RUnlock()

	if job == nil {
		return []string{}
	}

	job.mu.Lock()
	defer job.mu.Unlock()

	if offset >= len(job.Logs) {
		return []string{}
	}
	return job.Logs[offset:]
}
