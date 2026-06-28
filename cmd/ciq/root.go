package main

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type rootFlags struct {
	credsFile string
	context   string
	format    string
}

var rf rootFlags

func newRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "ciq",
		Short:         "AI-agent-friendly CI inspection CLI (read-only)",
		SilenceUsage:  true,
		SilenceErrors: false,
		Version:       version,
	}
	defaultCreds, _ := os.UserConfigDir()
	if defaultCreds != "" {
		defaultCreds += string(os.PathSeparator) + "ciq" + string(os.PathSeparator) + "credentials.yaml"
	}
	root.PersistentFlags().StringVar(&rf.credsFile, "credentials", defaultCreds, "path to credentials.yaml")
	root.PersistentFlags().StringVar(&rf.context, "context", "", "credentials context (defaults to default-context)")
	root.PersistentFlags().StringVar(&rf.format, "format", "", "output format: json|table|md (default: table on tty, json otherwise)")
	root.AddCommand(newJenkinsCmd())
	return root
}

func stdoutIsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}
