package api

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/aegisbox/aegisbox/internal/executor"
	"github.com/aegisbox/aegisbox/internal/logging"
)

const (
	APIVersion = "0.1.0"
)

// Handler handles HTTP requests for AegisBox API.
type Handler struct {
	mgr      *executor.ExecutionManager
	resolver executor.RuntimeResolver
	logger   *logging.Logger
}

// NewHandler creates a new Handler.
func NewHandler(mgr *executor.ExecutionManager, resolver executor.RuntimeResolver, logger *logging.Logger) *Handler {
	if logger == nil {
		logger = logging.Default()
	}
	return &Handler{
		mgr:      mgr,
		resolver: resolver,
		logger:   logger,
	}
}

// Health handles GET /health.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method Not Allowed"})
		return
	}

	var langs []string
	if h.resolver != nil {
		langs = h.resolver.SupportedLanguages()
	}

	res := HealthResponse{
		Status:             "ok",
		Version:            APIVersion,
		OS:                 runtime.GOOS,
		Arch:               runtime.GOARCH,
		SupportedLanguages: langs,
	}

	writeJSON(w, http.StatusOK, res)
}

// Execute handles POST /execute.
func (h *Handler) Execute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{Error: "Method Not Allowed"})
		return
	}

	// Limit request body to 128KB to prevent memory exhaustion on request payload
	r.Body = http.MaxBytesReader(w, r.Body, 128*1024)

	var payload ExecuteRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid JSON payload",
			Details: err.Error(),
		})
		return
	}

	domainReq := payload.ToDomain()
	if err := domainReq.Validate(); err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, ErrorResponse{
			Error:   "Validation Failed",
			Details: err.Error(),
		})
		return
	}

	// Context with overall deadline
	ctx, cancel := context.WithTimeout(r.Context(), domainReq.TimeoutDuration()+2*time.Second)
	defer cancel()

	result := h.mgr.Execute(ctx, domainReq)
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}
