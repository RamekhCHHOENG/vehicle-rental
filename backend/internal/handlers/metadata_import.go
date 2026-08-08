package handlers

import (
	"log"
	"net/http"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/ramekhchhoeng/car-rental/backend/internal/database"
	"github.com/ramekhchhoeng/car-rental/backend/internal/httputil"
)

// Importing the catalogue from NHTSA takes one request per make — around 1,900
// of them, several minutes end to end. That is far too long to hold an HTTP
// request open, so the import runs in the background and the admin screen polls
// this status instead.
//
// The status lives in memory. It describes one server's current activity, not
// something worth persisting: if the process restarts mid-import the work
// already committed stays committed, and re-running it is free because every
// insert is conflict-tolerant.
type importStatus struct {
	Running    bool                   `json:"running"`
	StartedAt  *time.Time             `json:"started_at,omitempty"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	WithModels bool                   `json:"with_models"`
	Result     *database.ImportResult `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

type ImportHandler struct {
	DB *gorm.DB

	mu     sync.Mutex
	status importStatus
}

// Status reports what the last or current import is doing.
// GET /api/admin/metadata/import
func (h *ImportHandler) Status(w http.ResponseWriter, r *http.Request) {
	h.mu.Lock()
	status := h.status
	h.mu.Unlock()

	httputil.JSON(w, http.StatusOK, status)
}

type importRequest struct {
	// Makes alone is two requests and finishes instantly, but leaves those makes
	// with no models — and a make with no models cannot be listed against, so it
	// is off by default only in the sense that the caller must ask for it.
	WithModels *bool `json:"with_models"`
}

// Start kicks off an import unless one is already running.
// POST /api/admin/metadata/import
func (h *ImportHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req importRequest
	// An empty body is a valid request: import everything.
	_ = decodeOptional(r, &req)
	withModels := boolOr(req.WithModels, true)

	h.mu.Lock()
	if h.status.Running {
		h.mu.Unlock()
		httputil.Error(w, http.StatusConflict, "an import is already running")
		return
	}
	started := time.Now()
	h.status = importStatus{Running: true, StartedAt: &started, WithModels: withModels}
	h.mu.Unlock()

	go h.run(withModels)

	httputil.JSON(w, http.StatusAccepted, map[string]any{
		"status": "started",
		"note":   "importing models takes several minutes; poll GET /api/admin/metadata/import",
	})
}

func (h *ImportHandler) run(withModels bool) {
	defer func() {
		// A panic in a background goroutine would take the whole API down, and
		// this one is parsing somebody else's JSON.
		if r := recover(); r != nil {
			log.Printf("nhtsa import panicked: %v", r)
			h.finish(nil, "the import failed unexpectedly")
		}
	}()

	result, err := database.ImportFromNHTSA(h.DB, withModels)
	if err != nil {
		h.finish(&result, err.Error())
		return
	}
	h.finish(&result, "")
}

func (h *ImportHandler) finish(result *database.ImportResult, errMessage string) {
	finished := time.Now()

	h.mu.Lock()
	defer h.mu.Unlock()

	h.status.Running = false
	h.status.FinishedAt = &finished
	h.status.Result = result
	h.status.Error = errMessage
}
