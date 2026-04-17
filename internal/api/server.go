package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/milad176/go-task-processor/internal/job"
)

type Server struct {
	repo   *job.Repository
	queue  chan job.Job
	server *http.Server
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewServer(repo *job.Repository, queue chan job.Job) *Server {
	s := &Server{
		repo:  repo,
		queue: queue,
	}

	mux := http.NewServeMux()

	mux.Handle("/jobs", LoggingMiddleware(http.HandlerFunc(s.handleCreateJob)))
	mux.Handle("/jobs/", LoggingMiddleware(http.HandlerFunc(s.handleGetJob)))

	s.server = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	return s
}

func (s *Server) Start() error {
	fmt.Println("🚀 Server listening on http://localhost:8080")
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var job job.Job

	// Decode request
	err := json.NewDecoder(r.Body).Decode(&job)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate
	err = job.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Assign system-controlled fields
	job.ID = uuid.New().String()
	job.Status = "pending"
	job.Retries = 0

	// Retry policy
	const defaultRetries = 5
	const maxAllowedRetries = 7

	if job.MaxRetries <= 0 {
		job.MaxRetries = defaultRetries
	} else if job.MaxRetries > maxAllowedRetries {
		job.MaxRetries = maxAllowedRetries
	}

	// Save once (clean)
	err = s.repo.Save(job)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save job")
		return
	}

	// Push to queue
	s.queue <- job

	// Return response
	writeJSON(w, http.StatusCreated, job)
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id := r.URL.Path[len("/jobs/"):]

	job, _ := s.repo.Get(id)

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
