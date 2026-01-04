package triton

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	endpoint string
	http     *http.Client
}

func NewClient(endpoint string, timeout time.Duration) (*Client, error) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return nil, fmt.Errorf("triton endpoint is required")
	}
	return &Client{
		endpoint: strings.TrimRight(endpoint, "/"),
		http: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

type inferenceInput struct {
	Name     string      `json:"name"`
	Datatype string      `json:"datatype"`
	Shape    []int64     `json:"shape"`
	Data     interface{} `json:"data"`
}

type inferenceOutput struct {
	Name string `json:"name"`
}

type inferenceRequest struct {
	Inputs  []inferenceInput  `json:"inputs"`
	Outputs []inferenceOutput `json:"outputs,omitempty"`
}

type inferenceResponse struct {
	Outputs []tensorResponse `json:"outputs"`
}

type tensorResponse struct {
	Name     string        `json:"name"`
	Datatype string        `json:"datatype"`
	Shape    []int64       `json:"shape"`
	Data     []interface{} `json:"data"`
}

func (c *Client) ModelReady(ctx context.Context, modelName string) error {
	url := fmt.Sprintf("%s/v2/models/%s/ready", c.endpoint, modelName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("triton model not ready: %s (%s)", modelName, strings.TrimSpace(string(body)))
	}
	return nil
}

func (c *Client) Infer(ctx context.Context, modelName string, inputs []inferenceInput, outputNames []string) (map[string]tensorResponse, error) {
	request := inferenceRequest{Inputs: inputs}
	if len(outputNames) > 0 {
		request.Outputs = make([]inferenceOutput, 0, len(outputNames))
		for _, name := range outputNames {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			request.Outputs = append(request.Outputs, inferenceOutput{Name: name})
		}
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v2/models/%s/infer", c.endpoint, modelName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("triton inference failed: %s", strings.TrimSpace(string(body)))
	}

	var decoded inferenceResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}

	outputs := make(map[string]tensorResponse, len(decoded.Outputs))
	for _, out := range decoded.Outputs {
		outputs[out.Name] = out
	}
	return outputs, nil
}
