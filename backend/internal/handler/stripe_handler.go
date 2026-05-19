package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/model"
	"crm0/backend/internal/service"
)

type StripeHandler struct {
	svc *service.StripeService
}

func NewStripeHandler(svc *service.StripeService) *StripeHandler {
	return &StripeHandler{svc: svc}
}

func (h *StripeHandler) GetBillingInfo(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	info, err := h.svc.GetBillingInfo(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (h *StripeHandler) CreateCheckout(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req model.CheckoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	sessionID, err := h.svc.CreateCheckoutSession(r.Context(), tenantID, req.PriceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"session_id": sessionID})
}

func (h *StripeHandler) CreatePortalSession(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	url, err := h.svc.CreatePortalSession(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"url": url})
}

func (h *StripeHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read body")
		return
	}

	var event struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook payload")
		return
	}

	data, _ := event.Data["object"].(map[string]any)
	if data == nil {
		data = event.Data
	}

	if err := h.svc.HandleWebhook(r.Context(), event.Type, data); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"received": true})
}
