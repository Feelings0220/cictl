package jenkins

import (
	"context"
	"fmt"
)

type QueueItem struct {
	ID    int64  `json:"id"`
	Why   string `json:"why"`
	Stuck bool   `json:"stuck"`
	Task  struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"task"`
	InQueueSince int64 `json:"inQueueSince"`
}

func (c *Client) ListQueue(ctx context.Context) ([]QueueItem, error) {
	var resp struct {
		Items []QueueItem `json:"items"`
	}
	if _, err := c.GET(ctx, "/queue/api/json", &resp); err != nil {
		return nil, fmt.Errorf("list queue: %w", err)
	}
	return resp.Items, nil
}
