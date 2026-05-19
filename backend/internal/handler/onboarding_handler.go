package handler

import (
	"encoding/json"
	"net/http"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/model"
	"crm0/backend/internal/service"

	"github.com/google/uuid"
)

type OnboardingHandler struct {
	svc *service.OnboardingService
}

func NewOnboardingHandler(svc *service.OnboardingService) *OnboardingHandler {
	return &OnboardingHandler{svc: svc}
}

func (h *OnboardingHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	status, err := h.svc.GetStatus(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *OnboardingHandler) CompleteStep(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	var req model.CompleteStepRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.CompleteStep(r.Context(), tenantID, req.Step); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	status, _ := h.svc.GetStatus(r.Context(), tenantID)
	writeJSON(w, http.StatusOK, status)
}

func (h *OnboardingHandler) SeedDemoData(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if err := h.svc.SeedDemoData(r.Context(), tenantID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	status, _ := h.svc.GetStatus(r.Context(), tenantID)
	writeJSON(w, http.StatusOK, status)
}

func (h *OnboardingHandler) Skip(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if err := h.svc.SkipOnboarding(r.Context(), tenantID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"skipped": true})
}

// OnboardingCheck is a public endpoint that returns onboarding status for a tenant.
func (h *OnboardingHandler) OnboardingCheck(w http.ResponseWriter, r *http.Request) {
	tenantIDStr := r.URL.Query().Get("tenant_id")
	if tenantIDStr == "" {
		writeError(w, http.StatusBadRequest, "tenant_id required")
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tenant_id")
		return
	}
	status, err := h.svc.GetStatus(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, status)
}
