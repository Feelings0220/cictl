package main

import (
	"context"
	"fmt"

	"github.com/kagent-dev/ciq/internal/config"
	"github.com/kagent-dev/ciq/internal/jenkins"
	"github.com/spf13/cobra"
)

func newJenkinsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "jenkins",
		Short: "Query Jenkins (read-only)",
	}
	c.AddCommand(newJenkinsWhoAmICmd())
	c.AddCommand(newJenkinsJobCmd())
	c.AddCommand(newJenkinsBuildCmd())
	return c
}

// newJenkinsClient resolves credentials and constructs a Jenkins client.
// Shared by all jenkins_*.go subcommands.
func newJenkinsClient(_ *cobra.Command) (*jenkins.Client, context.Context, error) {
	if rf.credsFile == "" {
		return nil, nil, fmt.Errorf("no credentials file (set --credentials or place credentials.yaml under your user config dir)")
	}
	creds, err := config.Load(rf.credsFile, rf.context)
	if err != nil {
		return nil, nil, err
	}
	cl, err := jenkins.New(creds)
	if err != nil {
		return nil, nil, err
	}
	return cl, context.Background(), nil
}
