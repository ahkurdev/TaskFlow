package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"taskflow/internal/middleware"
	"taskflow/internal/service"
)

type BoardHandler struct {
	boards *service.BoardService
}

func NewBoardHandler(boards *service.BoardService) *BoardHandler {
	return &BoardHandler{boards: boards}
}

type boardRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (h *BoardHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	var req boardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	board, err := h.boards.Create(r.Context(), userID, req.Name, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create board")
		return
	}

	writeJSON(w, http.StatusCreated, board)
}

func (h *BoardHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boards, err := h.boards.ListForUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list boards")
		return
	}

	writeJSON(w, http.StatusOK, boards)
}

func (h *BoardHandler) GetDetail(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}

	detail, err := h.boards.GetDetail(r.Context(), userID, boardID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "board not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load board")
		return
	}

	writeJSON(w, http.StatusOK, detail)
}

func (h *BoardHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}

	var req boardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	board, err := h.boards.Update(r.Context(), userID, boardID, req.Name, req.Description)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "board not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update board")
		return
	}

	writeJSON(w, http.StatusOK, board)
}

func (h *BoardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}

	if err := h.boards.Delete(r.Context(), userID, boardID); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			writeError(w, http.StatusNotFound, "board not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete board")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
