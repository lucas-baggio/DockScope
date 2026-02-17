package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/dockscope/dockscope/internal/usecase"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Server struct {
	listContainers        *usecase.ListContainers
	listImages            *usecase.ListImages
	listVolumes           *usecase.ListVolumes
	getSystemSummary      *usecase.GetSystemSummary
	streamContainerStats  *usecase.StreamContainerStats
	streamContainerLogs   *usecase.StreamContainerLogs
	executeContainerAction *usecase.ExecuteContainerAction
	log                   *slog.Logger
}

func NewServer(
	listContainers *usecase.ListContainers,
	listImages *usecase.ListImages,
	listVolumes *usecase.ListVolumes,
	getSystemSummary *usecase.GetSystemSummary,
	streamContainerStats *usecase.StreamContainerStats,
	streamContainerLogs *usecase.StreamContainerLogs,
	executeContainerAction *usecase.ExecuteContainerAction,
	log *slog.Logger,
) *Server {
	return &Server{
		listContainers:        listContainers,
		listImages:            listImages,
		listVolumes:           listVolumes,
		getSystemSummary:      getSystemSummary,
		streamContainerStats:  streamContainerStats,
		streamContainerLogs:   streamContainerLogs,
		executeContainerAction: executeContainerAction,
		log:                   log,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/containers", s.handleListContainers)
	mux.HandleFunc("POST /api/containers/", s.handleContainerAction)
	mux.HandleFunc("GET /api/system/summary", s.handleSystemSummary)
	mux.HandleFunc("GET /api/images", s.handleListImages)
	mux.HandleFunc("GET /api/volumes", s.handleListVolumes)
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/stats/{id}", s.handleStatsWebSocket)
	mux.HandleFunc("GET /api/logs/{id}", s.handleLogsWebSocket)
	return corsMiddleware(mux, s.log)
}

func (s *Server) handleSystemSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	out, err := s.getSystemSummary.Execute(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "api system summary failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to get system summary")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleListContainers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	all := r.URL.Query().Get("all") == "1" || r.URL.Query().Get("all") == "true"
	list, err := s.listContainers.Execute(ctx, usecase.ListContainersInput{All: all})
	if err != nil {
		s.log.ErrorContext(ctx, "api list containers failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to list containers")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type requestBodyContainerAction struct {
	Action string `json:"action"`
}

func (s *Server) handleContainerAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	suffix := strings.TrimPrefix(r.URL.Path, "/api/containers/")
	parts := strings.SplitN(suffix, "/", 2)
	if len(parts) != 2 || parts[1] != "action" {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	containerID := parts[0]
	if containerID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing container id")
		return
	}
	ctx := r.Context()
	var body requestBodyContainerAction
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	err := s.executeContainerAction.Execute(ctx, usecase.ExecuteContainerActionInput{
		ContainerID: containerID,
		Action:      body.Action,
	})
	if err != nil {
		msg := err.Error()
		if msg == "invalid action: must be one of start, stop, restart, pause, unpause" || msg == "missing container id" {
			writeJSONError(w, http.StatusBadRequest, msg)
			return
		}
		if strings.Contains(msg, "already") || strings.Contains(msg, "is not running") ||
			strings.Contains(msg, "is paused") || strings.Contains(msg, "No such container") {
			writeJSONError(w, http.StatusBadRequest, msg)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, msg)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (s *Server) handleListImages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.listImages.Execute(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "api list images failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to list images")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleListVolumes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	list, err := s.listVolumes.Execute(ctx)
	if err != nil {
		s.log.ErrorContext(ctx, "api list volumes failed", "error", err)
		writeJSONError(w, http.StatusInternalServerError, "failed to list volumes")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleStatsWebSocket(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")
	if containerID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing container id")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.WarnContext(r.Context(), "websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	conn.SetCloseHandler(func(code int, text string) error {
		cancel()
		return nil
	})

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	if err := s.streamContainerStats.Execute(ctx, containerID, conn); err != nil && ctx.Err() == nil {
		s.log.DebugContext(ctx, "stats stream ended", "container_id", containerID, "error", err)
	}
}

func (s *Server) handleLogsWebSocket(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")
	if containerID == "" {
		writeJSONError(w, http.StatusBadRequest, "missing container id")
		return
	}
	s.log.InfoContext(r.Context(), "logs websocket request", "path", r.URL.Path, "container_id", containerID)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.log.WarnContext(r.Context(), "websocket upgrade failed (logs)", "error", err)
		return
	}
	s.log.InfoContext(r.Context(), "logs websocket upgraded", "container_id", containerID)
	defer conn.Close()

	if err := conn.WriteMessage(websocket.TextMessage, []byte("")); err != nil {
		s.log.WarnContext(r.Context(), "logs websocket initial write failed", "error", err)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	conn.SetCloseHandler(func(code int, text string) error {
		cancel()
		return nil
	})

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	logWriter := &wsLogWriter{conn: conn}
	if err := s.streamContainerLogs.Execute(ctx, containerID, logWriter); err != nil && ctx.Err() == nil {
		s.log.DebugContext(ctx, "logs stream ended", "container_id", containerID, "error", err)
		if msg, _ := json.Marshal(map[string]string{"error": err.Error()}); len(msg) > 0 {
			_ = conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

type wsLogWriter struct{ conn *websocket.Conn }

func (w *wsLogWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
	err = w.conn.WriteMessage(websocket.TextMessage, p)
	return len(p), err
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func corsMiddleware(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	srv := &http.Server{Addr: addr, Handler: s.Handler()}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	s.log.InfoContext(ctx, "api server listening", "addr", addr)
	return srv.ListenAndServe()
}
