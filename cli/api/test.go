package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fsuhrau/automationhub/endpoints/api"
	"github.com/fsuhrau/automationhub/storage/models"
	"net/http"
)



func (c *Client) GetTests(ctx context.Context) ([]models.Test, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/tests", c.BaseURL), nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	var tests []models.Test
	if err := c.sendRequest(req, &tests); err != nil {
		return nil, err
	}

	return tests, nil
}

func (c *Client) ExecuteTest(ctx context.Context, testId int, appID int, params string) (*models.TestRun, error) {
	request := api.RunTestRequest{
		AppID: uint(appID),
		Params: params,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/test/%d/run", c.BaseURL, testId), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	var run models.TestRun
	if err := c.sendRequest(req, &run); err != nil {
		return nil, err
	}

	return &run, nil
}