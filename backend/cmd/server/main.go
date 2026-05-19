package main

import (
	"log"
	"net/http"

	"crm0/backend/internal/algorithm"
	"crm0/backend/internal/config"
	"crm0/backend/internal/handler"
	"crm0/backend/internal/middleware"
	"crm0/backend/internal/repository"
	"crm0/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := repository.NewDB(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Repositories
	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	customerRepo := repository.NewCustomerRepository(db)
	planRepo := repository.NewPlanRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	eventRepo := repository.NewEventRepository(db)
	algoRepo := repository.NewAlgorithmRepository(db)
	onboardingRepo := repository.NewOnboardingRepository(db)
	billingRepo := repository.NewBillingRepository(db)
	integrationRepo := repository.NewIntegrationRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	algoClient := algorithm.NewClient(cfg.AlgorithmServiceURL)

	// Services
	authSvc := service.NewAuthService(userRepo, tenantRepo, cfg)
	customerSvc := service.NewCustomerService(customerRepo, eventRepo, algoRepo, algoClient)
	planSvc := service.NewPlanService(planRepo)
	subscriptionSvc := service.NewSubscriptionService(subscriptionRepo)
	eventSvc := service.NewEventService(eventRepo)
	analyticsSvc := service.NewAnalyticsService(algoRepo, algoClient)
	onboardingSvc := service.NewOnboardingService(onboardingRepo, tenantRepo, customerRepo)
	stripeSvc := service.NewStripeService(billingRepo, tenantRepo, cfg)
	integrationSvc := service.NewIntegrationService(integrationRepo, customerRepo, orderRepo)
	orderSvc := service.NewOrderService(orderRepo)

	// Handlers
	authH := handler.NewAuthHandler(authSvc)
	customerH := handler.NewCustomerHandler(customerSvc)
	planH := handler.NewPlanHandler(planSvc)
	subscriptionH := handler.NewSubscriptionHandler(subscriptionSvc)
	eventH := handler.NewEventHandler(eventSvc)
	analyticsH := handler.NewAnalyticsHandler(analyticsSvc)
	onboardingH := handler.NewOnboardingHandler(onboardingSvc)
	stripeH := handler.NewStripeHandler(stripeSvc)
	integrationH := handler.NewIntegrationHandler(integrationSvc)
	orderH := handler.NewOrderHandler(orderSvc)

	mux := http.NewServeMux()

	// Auth (public)
	mux.HandleFunc("POST /api/v1/auth/login", authH.Login)
	mux.HandleFunc("POST /api/v1/auth/register", authH.Register)
	mux.HandleFunc("POST /api/v1/auth/refresh", authH.Refresh)

	// Customers
	mux.HandleFunc("GET /api/v1/customers", customerH.List)
	mux.HandleFunc("POST /api/v1/customers", customerH.Create)
	mux.HandleFunc("GET /api/v1/customers/{id}", customerH.Get)
	mux.HandleFunc("PUT /api/v1/customers/{id}", customerH.Update)
	mux.HandleFunc("DELETE /api/v1/customers/{id}", customerH.Delete)
	mux.HandleFunc("GET /api/v1/customers/{id}/insights", customerH.GetInsights)
	mux.HandleFunc("GET /api/v1/customers/{id}/events", customerH.GetEvents)

	// Subscriptions
	mux.HandleFunc("GET /api/v1/subscriptions", subscriptionH.List)
	mux.HandleFunc("POST /api/v1/subscriptions", subscriptionH.Create)
	mux.HandleFunc("GET /api/v1/subscriptions/{id}", subscriptionH.Get)
	mux.HandleFunc("PUT /api/v1/subscriptions/{id}", subscriptionH.Update)
	mux.HandleFunc("DELETE /api/v1/subscriptions/{id}", subscriptionH.Delete)
	mux.HandleFunc("GET /api/v1/subscriptions/metrics", subscriptionH.GetMetrics)

	// Plans
	mux.HandleFunc("GET /api/v1/plans", planH.List)
	mux.HandleFunc("POST /api/v1/plans", planH.Create)
	mux.HandleFunc("GET /api/v1/plans/{id}", planH.Get)
	mux.HandleFunc("PUT /api/v1/plans/{id}", planH.Update)
	mux.HandleFunc("DELETE /api/v1/plans/{id}", planH.Delete)

	// Events
	mux.HandleFunc("POST /api/v1/events", eventH.Create)
	mux.HandleFunc("GET /api/v1/events", eventH.GetRecent)

	// Analytics
	mux.HandleFunc("GET /api/v1/analytics/churn/predictions", analyticsH.GetChurnPredictions)
	mux.HandleFunc("POST /api/v1/analytics/churn/trigger-prediction", analyticsH.TriggerChurnPrediction)
	mux.HandleFunc("GET /api/v1/analytics/segments", analyticsH.GetSegments)
	mux.HandleFunc("POST /api/v1/analytics/segments/trigger-segmentation", analyticsH.TriggerSegmentation)
	mux.HandleFunc("GET /api/v1/analytics/ltv/predictions", analyticsH.GetLTVPredictions)
	mux.HandleFunc("GET /api/v1/analytics/ltv/channel-roi", analyticsH.GetChannelROI)
	mux.HandleFunc("GET /api/v1/analytics/nba/recommendations", analyticsH.GetNBARecommendations)
	mux.HandleFunc("POST /api/v1/analytics/nba/trigger-nba", analyticsH.TriggerNBA)
	mux.HandleFunc("GET /api/v1/dashboard", analyticsH.GetDashboard)

	// Onboarding
	mux.HandleFunc("GET /api/v1/onboarding/status", onboardingH.GetStatus)
	mux.HandleFunc("POST /api/v1/onboarding/complete-step", onboardingH.CompleteStep)
	mux.HandleFunc("POST /api/v1/onboarding/seed-demo", onboardingH.SeedDemoData)
	mux.HandleFunc("POST /api/v1/onboarding/skip", onboardingH.Skip)

	// Billing (Stripe)
	mux.HandleFunc("GET /api/v1/billing/info", stripeH.GetBillingInfo)
	mux.HandleFunc("POST /api/v1/billing/checkout", stripeH.CreateCheckout)
	mux.HandleFunc("POST /api/v1/billing/portal", stripeH.CreatePortalSession)
	mux.HandleFunc("POST /api/v1/billing/webhook", stripeH.HandleWebhook)

	// Platform Integrations
	mux.HandleFunc("GET /api/v1/integrations", integrationH.List)
	mux.HandleFunc("POST /api/v1/integrations/connect", integrationH.Connect)
	mux.HandleFunc("POST /api/v1/integrations/{platform}/disconnect", integrationH.Disconnect)
	mux.HandleFunc("POST /api/v1/integrations/{platform}/sync", integrationH.TriggerSync)

	// Orders
	mux.HandleFunc("GET /api/v1/orders", orderH.List)
	mux.HandleFunc("GET /api/v1/orders/{id}", orderH.Get)
	mux.HandleFunc("GET /api/v1/orders/metrics", orderH.GetMetrics)

	var h http.Handler = mux
	h = middleware.JWTAuth(cfg.JWTSecret)(h)
	h = middleware.Logging(h)
	h = middleware.CORS(h)

	log.Printf("Server starting on %s", cfg.AppAddr())
	if err := http.ListenAndServe(cfg.AppAddr(), h); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
