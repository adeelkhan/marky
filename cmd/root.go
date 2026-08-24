package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "marky",
		Short: "A markdown document manager powered by YAML",
	}
	root.AddCommand(newGenerateCmd())
	root.AddCommand(newViewCmd())
	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
