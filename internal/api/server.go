package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/milad176/go-task-processor/internal/job"
)

type Server struct {
	jobStore *job.JobStore
	queue    chan job.Job
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewServer(jobStore *job.JobStore, queue chan job.Job) *Server {
	return &Server{
		jobStore: jobStore,
		queue:    queue,
	}
}

func (s *Server) Start() {
	http.HandleFunc("/jobs", s.handleCreateJob)
	http.HandleFunc("/jobs/", s.handleGetJob)

	fmt.Println("🚀 Server listening on http://localhost:8080")

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var job job.Job

	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = job.Validate()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	job.ID = uuid.New().String()

	s.jobStore.Add(job) // Add job to store

	updated := s.jobStore.Get(job.ID) // get updated job from store

	s.queue <- updated // Push to queue

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(updated)
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Path[len("/jobs/"):]

	job := s.jobStore.Get(id)

	if job.ID == "" {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(job)
}
