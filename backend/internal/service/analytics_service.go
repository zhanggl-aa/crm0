package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"crm0/backend/internal/algorithm"
	"crm0/backend/internal/model"
	"crm0/backend/internal/repository"

	"github.com/google/uuid"
)

type AnalyticsService struct {
	algoRepo *repository.AlgorithmRepository
	algo     *algorithm.Client
}

func NewAnalyticsService(algoRepo *repository.AlgorithmRepository, algo *algorithm.Client) *AnalyticsService {
	return &AnalyticsService{algoRepo: algoRepo, algo: algo}
}

func (s *AnalyticsService) GetChurnPredictions(ctx context.Context, tenantID uuid.UUID) ([]*model.ChurnPrediction, error) {
	return s.algoRepo.GetChurnPredictions(ctx, tenantID)
}

func (s *AnalyticsService) TriggerChurnPrediction(ctx context.Context, tenantID uuid.UUID) error {
	result, err := s.algo.PredictChurnBatch(tenantID)
	if err != nil {
		return fmt.Errorf("churn prediction failed: %w", err)
	}
	predictions, ok := result["predictions"].([]any)
	if !ok {
		return nil
	}
	for _, p := range predictions {
		pm := p.(map[string]any)
		customerID, _ := uuid.Parse(pm["customer_id"].(string))
		factors, _ := json.Marshal(pm["factors"])
		pred := &model.ChurnPrediction{
			ID: uuid.New(), CustomerID: customerID, TenantID: tenantID, RiskScore: pm["risk_score"].(float64),
			RiskLevel: pm["risk_level"].(string), Factors: factors, PredictedAt: time.Now(),
		}
		s.algoRepo.SaveChurnPrediction(ctx, pred)
	}
	return nil
}

func (s *AnalyticsService) GetSegments(ctx context.Context, tenantID uuid.UUID) ([]*model.CustomerSegment, error) {
	return s.algoRepo.GetCustomerSegments(ctx, tenantID)
}

func (s *AnalyticsService) TriggerSegmentation(ctx context.Context, tenantID uuid.UUID, method string) error {
	result, err := s.algo.RunSegmentation(tenantID, method)
	if err != nil {
		return fmt.Errorf("segmentation failed: %w", err)
	}
	segments, ok := result["segments"].([]any)
	if !ok {
		return nil
	}
	for _, seg := range segments {
		sm := seg.(map[string]any)
		customerID, _ := uuid.Parse(sm["customer_id"].(string))
		cs := &model.CustomerSegment{
			ID: uuid.New(), CustomerID: customerID, TenantID: tenantID, SegmentType: sm["segment_type"].(string),
			SegmentName: sm["segment_name"].(string), Score: sm["score"].(float64), UpdatedAt: time.Now(),
		}
		s.algoRepo.SaveCustomerSegment(ctx, cs)
	}
	return nil
}

func (s *AnalyticsService) GetLTVPredictions(ctx context.Context, tenantID uuid.UUID) ([]*model.LTVPrediction, error) {
	return s.algoRepo.GetLTVPredictions(ctx, tenantID)
}

func (s *AnalyticsService) GetChannelROI(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	return s.algo.GetChannelROI(tenantID)
}

func (s *AnalyticsService) GetNBARecommendations(ctx context.Context, tenantID uuid.UUID) ([]*model.NBARecommendation, error) {
	return s.algoRepo.GetNBARecommendations(ctx, tenantID)
}

func (s *AnalyticsService) TriggerNBA(ctx context.Context, tenantID, customerID uuid.UUID) error {
	result, err := s.algo.GetNBA(tenantID, customerID)
	if err != nil {
		return fmt.Errorf("NBA failed: %w", err)
	}
	actions, ok := result["actions"].([]any)
	if !ok {
		return nil
	}
	for _, a := range actions {
		am := a.(map[string]any)
		custID, _ := uuid.Parse(am["customer_id"].(string))
		detail, _ := json.Marshal(am["action_detail"])
		rec := &model.NBARecommendation{
			ID: uuid.New(), CustomerID: custID, TenantID: tenantID, ActionType: am["action_type"].(string),
			ActionDetail: detail, ExpectedImpact: am["expected_impact"].(float64),
			Priority: int(am["priority"].(float64)), Status: "pending", CreatedAt: time.Now(),
		}
		s.algoRepo.SaveNBARecommendation(ctx, rec)
	}
	return nil
}

func (s *AnalyticsService) GetDashboard(ctx context.Context, tenantID uuid.UUID) (*model.DashboardOverview, error) {
	return s.algoRepo.GetDashboardOverview(ctx, tenantID)
}
