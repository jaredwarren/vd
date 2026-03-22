package server

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jaredwarren/ytdl/internal/jobs"
)

// API serves JSON endpoints.
type API struct {
	store *jobs.Store
}

// NewAPI returns handlers for /api/*.
func NewAPI(store *jobs.Store) *API {
	return &API{store: store}
}

type createJobReq struct {
	URL string `json:"url"`
}

type createJobResp struct {
	ID     string      `json:"id"`
	Status jobs.Status `json:"status"`
}

type errResp struct {
	Error string `json:"error"`
}

// Register attaches routes to mux (Go 1.22+ patterns).
func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/jobs", a.handleCreateJob)
	mux.HandleFunc("GET /api/jobs/{id}", a.handleGetJob)
}

func (a *API) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	var req createJobReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "invalid JSON body"})
		return
	}
	raw := strings.TrimSpace(req.URL)
	if raw == "" {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "url is required"})
		return
	}
	parsed, err := url.Parse(raw)
	if err != nil || !isYouTubeURL(parsed) {
		writeJSON(w, http.StatusBadRequest, errResp{Error: "only YouTube URLs are allowed"})
		return
	}
	id, err := randomID()
	if err != nil {
		log.Printf("job id: %v", err)
		writeJSON(w, http.StatusInternalServerError, errResp{Error: "could not create job"})
		return
	}
	job, err := a.store.Create(id, raw)
	if err != nil {
		writeJSON(w, http.StatusServiceUnavailable, errResp{Error: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createJobResp{ID: job.ID, Status: job.Status})
}

func (a *API) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeJSON(w, http.StatusNotFound, errResp{Error: "not found"})
		return
	}
	job, ok := a.store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, errResp{Error: "not found"})
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func isYouTubeURL(u *url.URL) bool {
	if u == nil {
		return false
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	switch host {
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com",
		"youtu.be", "www.youtu.be":
		return true
	default:
		return strings.HasSuffix(host, ".youtube.com")
	}
}

// LogMiddleware wraps a handler with request logging.
func LogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
