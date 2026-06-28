package jenkins

import (
	"context"
	"fmt"
)

type JobSummary struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

type HealthReport struct {
	Description string `json:"description"`
	Score       int    `json:"score"`
}

type BuildRef struct {
	Number int    `json:"number"`
	URL    string `json:"url"`
}

type Job struct {
	Name               string         `json:"name"`
	FullName           string         `json:"fullName"`
	URL                string         `json:"url"`
	Description        string         `json:"description"`
	Buildable          bool           `json:"buildable"`
	Color              string         `json:"color"`
	InQueue            bool           `json:"inQueue"`
	HealthReport       []HealthReport `json:"healthReport"`
	Builds             []BuildRef     `json:"builds"`
	LastBuild          *BuildRef      `json:"lastBuild"`
	LastCompletedBuild *BuildRef      `json:"lastCompletedBuild"`
	LastFailedBuild    *BuildRef      `json:"lastFailedBuild"`
}

func (c *Client) ListJobs(ctx context.Context, folder string) ([]JobSummary, error) {
	path := JobPath(folder) + "/api/json?tree=jobs[name,url,color]"
	if folder == "" {
		path = "/api/json?tree=jobs[name,url,color]"
	}
	var resp struct {
		Jobs []JobSummary `json:"jobs"`
	}
	if _, err := c.GET(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	return resp.Jobs, nil
}

func (c *Client) GetJob(ctx context.Context, name string) (Job, error) {
	var j Job
	if _, err := c.GET(ctx, JobPath(name)+"/api/json", &j); err != nil {
		return Job{}, fmt.Errorf("get job %s: %w", name, err)
	}
	return j, nil
}

func (c *Client) GetJobConfig(ctx context.Context, name string) ([]byte, error) {
	body, err := c.GET(ctx, JobPath(name)+"/config.xml", nil)
	if err != nil {
		return nil, fmt.Errorf("get job config %s: %w", name, err)
	}
	return body, nil
}
