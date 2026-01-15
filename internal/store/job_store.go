package store

import (
	"log"
	"sync"

	"github.com/changty97/macvmorx/internal/models"
)

// JobStore manages the state of GitHub workflow jobs.
type JobStore struct {
	jobs sync.Map // Stores map[int64]*models.JobStatus, using sync.Map for concurrent access
}

// NewJobStore creates and initializes a new JobStore.
func NewJobStore() *JobStore {
	return &JobStore{}
}

// AddJob adds a new job or updates an existing one.
func (s *JobStore) AddJob(job models.JobStatus) {
	s.jobs.Store(job.JobID, &job)
	log.Printf("Job %d (%s) added/updated with status: %s", job.JobID, job.RunnerName, job.Status)
}

// UpdateJobStatus updates the status and other fields of an existing job.
func (s *JobStore) UpdateJobStatus(jobID int64, updateFn func(*models.JobStatus)) {
	if val, ok := s.jobs.Load(jobID); ok {
		job := val.(*models.JobStatus)
		updateFn(job)
		s.jobs.Store(jobID, job) // Store the updated job
		log.Printf("Job %d (%s) status updated to: %s", job.JobID, job.RunnerName, job.Status)
	} else {
		log.Printf("Warning: Attempted to update non-existent job with ID: %d", jobID)
	}
}

// GetJobStatus retrieves the status of a specific job.
func (s *JobStore) GetJobStatus(jobID int64) (*models.JobStatus, bool) {
	if val, ok := s.jobs.Load(jobID); ok {
		return val.(*models.JobStatus), true
	}
	return nil, false
}

// GetAllJobs retrieves the statuses of all jobs.
func (s *JobStore) GetAllJobs() []*models.JobStatus {
	var allJobs []*models.JobStatus
	s.jobs.Range(func(key, value interface{}) bool {
		job := value.(*models.JobStatus)
		allJobs = append(allJobs, job)
		return true
	})
	return allJobs
}
