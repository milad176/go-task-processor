package job

import "sync"

type JobStore struct {
	mu   sync.Mutex
	jobs map[string]Job
}

func NewJobStore() *JobStore {
	return &JobStore{
		jobs: make(map[string]Job),
	}
}

func (s *JobStore) Add(job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	job.Status = "pending"
	s.jobs[job.ID] = job
}

func (s *JobStore) Get(id string) Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.jobs[id]
}

func (s *JobStore) UpdateStatus(id string, status string) Job {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := s.jobs[id]
	job.Status = status
	s.jobs[id] = job

	return job
}
