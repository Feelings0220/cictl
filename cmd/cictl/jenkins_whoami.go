package main

import (
	"os"

	"github.com/Feelings0220/cictl/internal/output"
	"github.com/spf13/cobra"
)

func newJenkinsWhoAmICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show the authenticated user",
		RunE: func(cmd *cobra.Command, args []string) error {
			cl, ctx, err := newJenkinsClient(cmd)
			if err != nil {
				return err
			}
			me, err := cl.WhoAmI(ctx)
			if err != nil {
				return err
			}
			rows := []map[string]any{{"id": me.ID, "fullName": me.FullName, "authenticated": me.Authenticated}}
			return output.Render(os.Stdout, output.Detect(stdoutIsTTY(), rf.format), rows)
		},
	}
}
