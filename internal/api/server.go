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

func (s *Server) Start() error {
	http.Handle("/jobs", LoggingMiddleware(http.HandlerFunc(s.handleCreateJob)))
	http.Handle("/jobs/", LoggingMiddleware(http.HandlerFunc(s.handleGetJob)))

	fmt.Println("🚀 Server listening on http://localhost:8080")

	return http.ListenAndServe(":8080", nil)
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var job job.Job

	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = job.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
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
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := r.URL.Path[len("/jobs/"):]

	job := s.jobStore.Get(id)

	if job.ID == "" {
		writeError(w, http.StatusNotFound, "Job not found")
		return
	}

	json.NewEncoder(w).Encode(job)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
