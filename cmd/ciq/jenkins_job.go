package main

import (
	"fmt"
	"os"

	"github.com/kagent-dev/ciq/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsJobCmd() *cobra.Command {
	c := &cobra.Command{Use: "job", Short: "Inspect Jenkins jobs"}

	listFolder := ""
	list := &cobra.Command{
		Use:   "list",
		Short: "List jobs (optionally inside a folder)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			jobs, err := cl.ListJobs(ctx, listFolder)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(jobs))
			for _, j := range jobs {
				rows = append(rows, map[string]any{"name": j.Name, "color": j.Color, "url": j.URL})
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}
	list.Flags().StringVar(&listFolder, "folder", "", "folder path (e.g. team/service)")

	get := &cobra.Command{
		Use:   "get <name>",
		Short: "Get a job's metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			j, err := cl.GetJob(ctx, args[0])
			if err != nil {
				return err
			}
			// For non-JSON formats flatten to a single-row table.
			f := output.Detect(stdoutIsTTY(), rf.format)
			if f == output.FormatJSON {
				return output.Render(os.Stdout, f, j)
			}
			row := map[string]any{"name": j.Name, "color": j.Color, "buildable": j.Buildable, "inQueue": j.InQueue}
			if j.LastBuild != nil {
				row["lastBuild"] = j.LastBuild.Number
			}
			return output.Render(os.Stdout, f, []map[string]any{row})
		},
	}

	cfg := &cobra.Command{
		Use:   "config <name>",
		Short: "Print a job's config.xml",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			b, err := cl.GetJobConfig(ctx, args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(os.Stdout, string(b))
			return err
		},
	}

	c.AddCommand(list, get, cfg)
	return c
}
