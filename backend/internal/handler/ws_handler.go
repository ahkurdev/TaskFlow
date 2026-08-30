package handler

import (
	"errors"
	"net/http"

	"taskflow/internal/service"
	"taskflow/internal/ws"
)

type WSHandler struct {
	hub    *ws.Hub
	boards *service.BoardService
	auth   *service.AuthService
}

func NewWSHandler(hub *ws.Hub, boards *service.BoardService, auth *service.AuthService) *WSHandler {
	return &WSHandler{hub: hub, boards: boards, auth: auth}
}

// Serve upgrades the connection to a WebSocket and subscribes it to live
// updates for the given board. Auth is passed via a "token" query param
// (the browser WebSocket API cannot set custom headers).
func (h *WSHandler) Serve(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	userID, err := h.auth.ParseToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}

	if err := h.boards.VerifyOwnership(r.Context(), userID, boardID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "board not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to verify board", http.StatusInternalServerError)
		return
	}

	_ = ws.Upgrade(h.hub, w, r, boardID)
}
