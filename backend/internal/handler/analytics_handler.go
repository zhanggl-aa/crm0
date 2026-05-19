package handler

import (
	"net/http"

	"crm0/backend/internal/middleware"
	"crm0/backend/internal/service"

	"github.com/google/uuid"
)

type AnalyticsHandler struct {
	svc *service.AnalyticsService
}

func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

func (h *AnalyticsHandler) GetChurnPredictions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	predictions, err := h.svc.GetChurnPredictions(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, predictions)
}

func (h *AnalyticsHandler) TriggerChurnPrediction(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	if err := h.svc.TriggerChurnPrediction(r.Context(), tenantID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *AnalyticsHandler) GetSegments(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	segments, err := h.svc.GetSegments(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, segments)
}

func (h *AnalyticsHandler) TriggerSegmentation(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	method := r.URL.Query().Get("method")
	if method == "" {
		method = "rfm"
	}
	if err := h.svc.TriggerSegmentation(r.Context(), tenantID, method); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *AnalyticsHandler) GetLTVPredictions(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	predictions, err := h.svc.GetLTVPredictions(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, predictions)
}

func (h *AnalyticsHandler) GetChannelROI(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	roi, err := h.svc.GetChannelROI(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, roi)
}

func (h *AnalyticsHandler) GetNBARecommendations(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	recs, err := h.svc.GetNBARecommendations(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, recs)
}

func (h *AnalyticsHandler) TriggerNBA(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	customerIDStr := r.URL.Query().Get("customer_id")
	var customerID uuid.UUID
	if customerIDStr != "" {
		var err error
		customerID, err = uuid.Parse(customerIDStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid customer_id")
			return
		}
	}
	if err := h.svc.TriggerNBA(r.Context(), tenantID, customerID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func (h *AnalyticsHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.GetTenantID(r.Context())
	dashboard, err := h.svc.GetDashboard(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}
