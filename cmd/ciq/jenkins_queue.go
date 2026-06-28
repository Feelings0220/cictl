package main

import (
	"os"

	"github.com/kagent-dev/ciq/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsQueueCmd() *cobra.Command {
	c := &cobra.Command{Use: "queue", Short: "Inspect Jenkins build queue"}
	list := &cobra.Command{
		Use: "list", Short: "List queued items",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			items, err := cl.ListQueue(ctx)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(items))
			for _, q := range items {
				rows = append(rows, map[string]any{
					"id": q.ID, "task": q.Task.Name, "why": q.Why, "stuck": q.Stuck,
				})
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}
	c.AddCommand(list)
	return c
}
