package main

import (
	"fmt"
	"os"

	"github.com/kagent-dev/ciq/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsCloudCmd() *cobra.Command {
	c := &cobra.Command{Use: "cloud", Short: "Inspect Jenkins clouds (K8s/EC2/etc.)"}

	list := &cobra.Command{
		Use: "list", Short: "List configured clouds",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			cs, err := cl.ListClouds(ctx)
			if err != nil {
				return err
			}
			rows := make([]map[string]any, 0, len(cs))
			for _, x := range cs {
				rows = append(rows, map[string]any{"name": x.Name, "kind": x.Kind})
			}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}

	get := &cobra.Command{
		Use: "get <name>", Short: "Print a cloud's config.xml",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			b, err := cl.GetCloudConfig(ctx, args[0])
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(os.Stdout, string(b))
			return err
		},
	}

	c.AddCommand(list, get)
	return c
}
