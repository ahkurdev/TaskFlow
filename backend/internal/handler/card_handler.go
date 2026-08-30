package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"taskflow/internal/middleware"
	"taskflow/internal/service"
)

type CardHandler struct {
	cards *service.CardService
}

func NewCardHandler(cards *service.CardService) *CardHandler {
	return &CardHandler{cards: cards}
}

type cardRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type moveCardRequest struct {
	TargetListID int64 `json:"target_list_id"`
	Position     int   `json:"position"`
}

func (h *CardHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}
	listID, ok := parseIDParam(w, r, "listID")
	if !ok {
		return
	}

	var req cardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	card, err := h.cards.Create(r.Context(), userID, boardID, listID, req.Title, req.Description)
	if err != nil {
		writeCardError(w, err, "failed to create card")
		return
	}

	writeJSON(w, http.StatusCreated, card)
}

func (h *CardHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}
	listID, ok := parseIDParam(w, r, "listID")
	if !ok {
		return
	}
	cardID, ok := parseIDParam(w, r, "cardID")
	if !ok {
		return
	}

	var req cardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	card, err := h.cards.Update(r.Context(), userID, boardID, listID, cardID, req.Title, req.Description)
	if err != nil {
		writeCardError(w, err, "failed to update card")
		return
	}

	writeJSON(w, http.StatusOK, card)
}

func (h *CardHandler) Move(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}
	listID, ok := parseIDParam(w, r, "listID")
	if !ok {
		return
	}
	cardID, ok := parseIDParam(w, r, "cardID")
	if !ok {
		return
	}

	var req moveCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	card, err := h.cards.Move(r.Context(), userID, boardID, listID, cardID, req.TargetListID, req.Position)
	if err != nil {
		writeCardError(w, err, "failed to move card")
		return
	}

	writeJSON(w, http.StatusOK, card)
}

func (h *CardHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	boardID, ok := parseIDParam(w, r, "boardID")
	if !ok {
		return
	}
	listID, ok := parseIDParam(w, r, "listID")
	if !ok {
		return
	}
	cardID, ok := parseIDParam(w, r, "cardID")
	if !ok {
		return
	}

	if err := h.cards.Delete(r.Context(), userID, boardID, listID, cardID); err != nil {
		writeCardError(w, err, "failed to delete card")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeCardError(w http.ResponseWriter, err error, fallback string) {
	if errors.Is(err, service.ErrNotFound) {
		writeError(w, http.StatusNotFound, "card, list or board not found")
		return
	}
	writeError(w, http.StatusInternalServerError, fallback)
}
