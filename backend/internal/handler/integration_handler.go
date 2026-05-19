package handler

import (
	"encoding/json"
	"net/http"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/model"
	"crm0/backend/internal/service"
)

type IntegrationHandler struct {
	svc *service.IntegrationService
}

func NewIntegrationHandler(svc *service.IntegrationService) *IntegrationHandler {
	return &IntegrationHandler{svc: svc}
}

func (h *IntegrationHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	integrations, err := h.svc.List(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": integrations})
}

func (h *IntegrationHandler) Connect(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req model.ConnectPlatformRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Platform == "" || req.Code == "" {
		writeError(w, http.StatusBadRequest, "platform and code are required")
		return
	}
	result, err := h.svc.Connect(r.Context(), tenantID, req.Platform, req.Code)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *IntegrationHandler) Disconnect(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	platform := r.PathValue("platform")
	if platform == "" {
		writeError(w, http.StatusBadRequest, "platform is required")
		return
	}
	if err := h.svc.Disconnect(r.Context(), tenantID, platform); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"disconnected": true})
}

func (h *IntegrationHandler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	platform := r.PathValue("platform")
	if platform == "" {
		writeError(w, http.StatusBadRequest, "platform is required")
		return
	}
	if err := h.svc.TriggerSync(r.Context(), tenantID, platform); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"syncing": true})
}
