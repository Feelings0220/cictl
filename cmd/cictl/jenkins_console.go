package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Feelings0220/cictl/internal/jenkins"
	"github.com/spf13/cobra"
)

func newJenkinsConsoleCmd() *cobra.Command {
	var tail int
	var full bool
	c := &cobra.Command{
		Use:   "console <job> <number>",
		Short: "Print build console log (default: last 200 lines)",
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
			body, err := cl.GetConsole(ctx, args[0], n)
			if err != nil {
				return err
			}
			out := body
			if !full {
				out = jenkins.LastNLines(body, tail)
				if len(out) < len(body) {
					fmt.Fprintf(os.Stderr, "(showing last %d lines of %d bytes; use --full for everything)\n", tail, len(body))
				}
			}
			_, err = os.Stdout.Write(out)
			return err
		},
	}
	c.Flags().IntVar(&tail, "tail", 200, "show last N lines")
	c.Flags().BoolVar(&full, "full", false, "print the entire console log")
	return c
}
