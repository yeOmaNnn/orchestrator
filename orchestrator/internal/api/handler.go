package api

import (
	"encoding/json"
	"net/http"
	
	"strings"
	"github.com/google/uuid"

	"github.com/yeOmaNnn/orchestrator/internal/domain"
	"github.com/yeOmaNnn/orchestrator/internal/engine"
	"github.com/yeOmaNnn/orchestrator/internal/storage"
)

type Handler struct {
	engine *engine.Engine
	taskRepo storage.TaskRepository
	stepRepo storage.StepRepository
}

func NewHandler(
	engine *engine.Engine, 
	taskRepo storage.TaskRepository, 
	stepRepo storage.StepRepository,
) *Handler {
	return &Handler{
		engine: engine,
		taskRepo: taskRepo,
		stepRepo: stepRepo,
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/tasks", h.createTask)
	mux.HandleFunc("/tasks/", h.handleTaskByID)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Goal string `json:"goal"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	task := domain.Task {
		ID: uuid.New(), 
		Goal: req.Goal, 
		Status: domain.TaskPending,
	}

	if err := h.taskRepo.Create(r.Context(), &task); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.engine.InitTaskExecution(r.Context(), task); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go h.engine.RunTaskLoop(r.Context(), task.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)

}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Path[len("/tasks/"):]
	taskID, err := uuid.Parse(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return 
	}

	task, err := h.taskRepo.GetByID(r.Context(), taskID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	steps, _ := h.stepRepo.GetByTask(r.Context(), taskID)

	resp := struct {
		Task *domain.Task `json:"task"`
		Steps []domain.Step `json:"steps"`
	} {
		Task: task, 
		Steps: steps,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleTaskByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}

	taskID, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, "invalid task id", http.StatusBadRequest)
		return
	}

	action := parts[1]

	switch action {
	case "cancel":
		h.handleCancelTask(w, r, taskID)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) handleCancelTask(
	w http.ResponseWriter,
	r *http.Request,
	taskID uuid.UUID,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	if err := h.engine.CancelTask(ctx, taskID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "cancelled",
	})
}

func (h *Handler) cancelTask(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := chi.URLParam(r, "id")

	taskID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid id", 400)
		return
	}

	if err := h.engine.CancelTask(r.Context(), taskID); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
