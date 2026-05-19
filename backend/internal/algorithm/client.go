package algorithm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) post(path string, body any) (map[string]any, error) {
	data, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("algorithm service unreachable: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("algorithm service error: %s", string(raw))
	}

	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *Client) PredictChurn(tenantID, customerID uuid.UUID) (map[string]any, error) {
	return c.post("/churn/predict", map[string]string{
		"tenant_id":   tenantID.String(),
		"customer_id": customerID.String(),
	})
}

func (c *Client) PredictChurnBatch(tenantID uuid.UUID) (map[string]any, error) {
	return c.post("/churn/batch", map[string]string{
		"tenant_id": tenantID.String(),
	})
}

func (c *Client) PredictLTV(tenantID, customerID uuid.UUID) (map[string]any, error) {
	return c.post("/ltv/predict", map[string]string{
		"tenant_id":   tenantID.String(),
		"customer_id": customerID.String(),
	})
}

func (c *Client) GetChannelROI(tenantID uuid.UUID) (map[string]any, error) {
	return c.post("/ltv/channel-roi", map[string]string{
		"tenant_id": tenantID.String(),
	})
}

func (c *Client) RunSegmentation(tenantID uuid.UUID, method string) (map[string]any, error) {
	return c.post("/segments/run", map[string]string{
		"tenant_id":     tenantID.String(),
		"segment_type":  method,
	})
}

func (c *Client) GetNBA(tenantID, customerID uuid.UUID) (map[string]any, error) {
	return c.post("/nba/recommend", map[string]string{
		"tenant_id":   tenantID.String(),
		"customer_id": customerID.String(),
	})
}
