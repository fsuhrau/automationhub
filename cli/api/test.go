package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/fsuhrau/automationhub/endpoints/api"
	"github.com/fsuhrau/automationhub/storage/models"
)

func (c *Client) GetTests(ctx context.Context) ([]models.Test, error) {
	var tests []models.Test
	if err := c.sendJSON(ctx, jsonCall{method: http.MethodGet, path: "/tests", out: &tests}); err != nil {
		return nil, err
	}
	return tests, nil
}

func (c *Client) ExecuteTest(ctx context.Context, testID uint, request api.RunTestRequest) (*models.TestRun, error) {
	var run models.TestRun
	call := jsonCall{http.MethodPost, fmt.Sprintf("/test/%d/run", testID), request, &run}
	if err := c.sendJSON(ctx, call); err != nil {
		return nil, err
	}
	return &run, nil
}
