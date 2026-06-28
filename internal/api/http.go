package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/yusnelgg/agent-bridge/internal/nats"
	"github.com/yusnelgg/agent-bridge/internal/protocol"
	"github.com/yusnelgg/agent-bridge/internal/store"
	"github.com/yusnelgg/agent-bridge/internal/web"
)

type HTTPServer struct {
	store    *store.Store
	nats     *nats.Client
	identity string
	hub      *WSHub
}

func NewHTTPServer(addr, identity string, s *store.Store, nc *nats.Client, hub *WSHub) *http.Server {
	h := &HTTPServer{store: s, nats: nc, identity: identity, hub: hub}

	mux := http.NewServeMux()
	mux.HandleFunc("/messages/new", h.handleGetMessages)
	mux.HandleFunc("/messages/send", h.handleSendMessage)
	mux.HandleFunc("/messages/read", h.handleMarkRead)
	mux.HandleFunc("/tasks/delegate", h.handleDelegateTask)
	mux.HandleFunc("/tasks/list", h.handleListTasks)
	mux.HandleFunc("/tasks/status", h.handleTaskStatus)
	mux.HandleFunc("/context/share", h.handleShareContext)
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/messages/all", h.handleGetAllMessages)
	mux.HandleFunc("/ws", h.hub.HandleWS)
	mux.HandleFunc("/", web.DashboardHandler)

	return &http.Server{
		Addr:    addr,
		Handler: withCORS(mux),
	}
}

func (h *HTTPServer) Hub() *WSHub {
	return h.hub
}

func (h *HTTPServer) handleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	unread := r.URL.Query().Get("unread") == "true"
	msgs, err := h.store.GetMessages(h.identity, unread)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, msgs)
}

func (h *HTTPServer) handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		To      string   `json:"to"`
		Type    string   `json:"type"`
		Content string   `json:"content"`
		TaskID  string   `json:"task_id,omitempty"`
		Files   []string `json:"files,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	msg := &protocol.Message{
		ID:        uuid.New().String(),
		From:      h.identity,
		To:        req.To,
		Type:      protocol.MessageType(req.Type),
		Content:   req.Content,
		TaskID:    req.TaskID,
		Files:     req.Files,
		CreatedAt: time.Now(),
		Read:      false,
	}

	targetSubject := "agent." + req.To
	if err := h.nats.SendMessage(msg, targetSubject); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(msg)

	writeJSON(w, map[string]string{"status": "sent", "id": msg.ID})
}

func (h *HTTPServer) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if err := h.store.MarkRead(req.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func (h *HTTPServer) handleDelegateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		To          string `json:"to"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	now := time.Now()

	task := &protocol.Task{
		ID:          uuid.New().String(),
		From:        h.identity,
		To:          req.To,
		Description: req.Description,
		Status:      protocol.TaskPending,
		CreatedAt:   now,
	}

	if err := h.store.SaveTask(task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msg := &protocol.Message{
		ID:      uuid.New().String(),
		From:    h.identity,
		CreatedAt: now,
		To:      req.To,
		Type:    protocol.TypeTaskDelegate,
		Content: req.Description,
		TaskID:  task.ID,
		Read:    false,
	}

	targetSubject := "agent." + req.To
	if err := h.nats.SendMessage(msg, targetSubject); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(msg)

	writeJSON(w, map[string]string{"status": "delegated", "task_id": task.ID})
}

func (h *HTTPServer) handleListTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	activeOnly := r.URL.Query().Get("active") == "true"
	tasks, err := h.store.GetTasks(h.identity, activeOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, tasks)
}

func (h *HTTPServer) handleTaskStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
		Result string `json:"result"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	task, err := h.store.GetTaskByID(req.TaskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if task == nil {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	if err := h.store.UpdateTaskStatus(req.TaskID, protocol.TaskStatus(req.Status), req.Result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if req.Status == string(protocol.TaskDone) || req.Status == string(protocol.TaskFailed) {
		msg := &protocol.Message{
			ID:      uuid.New().String(),
			From:    h.identity,
			To:      task.From,
			Type:    protocol.TypeTaskResult,
			Content: req.Result,
			TaskID:  req.TaskID,
			CreatedAt: time.Now(),
		}
		targetSubject := "agent." + task.From
		if err := h.nats.SendMessage(msg, targetSubject); err != nil {
			log.Printf("[http] error notificando resultado: %v", err)
		}
		h.hub.Broadcast(msg)
	}

	writeJSON(w, map[string]string{"status": "updated"})
}

type shareRequest struct {
	To      string   `json:"to"`
	Content string   `json:"content"`
	Files   []string `json:"files"`
}

func (h *HTTPServer) handleShareContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req shareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	msg := &protocol.Message{
		ID:      uuid.New().String(),
		From:    h.identity,
		To:      req.To,
		Type:    protocol.TypeShareContext,
		Content: req.Content,
		Files:   req.Files,
		Read:    false,
	}

	targetSubject := "agent." + req.To
	if err := h.nats.SendMessage(msg, targetSubject); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.hub.Broadcast(msg)

	writeJSON(w, map[string]string{"status": "shared"})
}

func (h *HTTPServer) handleGetAllMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	msgs, err := h.store.GetAllMessages(h.identity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{
		"identity": h.identity,
		"messages": msgs,
	})
}

func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]string{
		"identity": h.identity,
		"status":   "ok",
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
