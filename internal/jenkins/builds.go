package jenkins

import (
	"context"
	"fmt"
)

type Build struct {
	Number      int    `json:"number"`
	Result      string `json:"result"`
	Building    bool   `json:"building"`
	Duration    int64  `json:"duration"`
	Timestamp   int64  `json:"timestamp"`
	URL         string `json:"url"`
	DisplayName string `json:"displayName"`
}

func (c *Client) ListBuilds(ctx context.Context, jobName string, limit int) ([]Build, error) {
	if limit <= 0 {
		limit = 20
	}
	path := fmt.Sprintf("%s/api/json?tree=builds[number,result,building,duration,timestamp,url,displayName]{0,%d}",
		JobPath(jobName), limit)
	var resp struct {
		Builds []Build `json:"builds"`
	}
	if _, err := c.GET(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("list builds for %s: %w", jobName, err)
	}
	return resp.Builds, nil
}

func (c *Client) GetBuild(ctx context.Context, jobName string, number int) (Build, error) {
	path := fmt.Sprintf("%s/%d/api/json", JobPath(jobName), number)
	var b Build
	if _, err := c.GET(ctx, path, &b); err != nil {
		return Build{}, fmt.Errorf("get build %s #%d: %w", jobName, number, err)
	}
	return b, nil
}
