package main

import (
	"os"
	"strconv"

	"github.com/Feelings0220/cictl/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsBuildCmd() *cobra.Command {
	c := &cobra.Command{Use: "build", Short: "Inspect Jenkins builds"}

	var limit int
	list := &cobra.Command{
		Use:   "list <job>",
		Short: "List recent builds for a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			bs, err := cl.ListBuilds(ctx, args[0], limit)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(bs))
			for _, b := range bs {
				rows = append(rows, map[string]any{
					"number": b.Number, "result": b.Result, "building": b.Building,
					"duration_ms": b.Duration, "url": b.URL,
				})
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}
	list.Flags().IntVar(&limit, "limit", 20, "max number of builds to list")

	get := &cobra.Command{
		Use:   "get <job> <number>",
		Short: "Get build metadata",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			b, err := cl.GetBuild(ctx, args[0], n)
			if err != nil {
				return err
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), b)
		},
	}

	c.AddCommand(list, get)
	return c
}
