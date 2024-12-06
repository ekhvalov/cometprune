package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ekhvalov/cometprune/internal/pruner"
)

var (
	keepBlocks int64
	path       string

	// rootCmd represents a new command
	rootCmd = &cobra.Command{
		Use:   "cometprune",
		Short: "Prunes CometBFT data",
		Long:  `Prunes CometBFT data and retaining only a specified number of recent blocks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pruner.Prune(path, keepBlocks)
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().Int64VarP(&keepBlocks, "keep-blocks", "k", 10, "Specify the number of blocks to keep")
	rootCmd.Flags().StringVarP(&path, "path", "p", "./data", "Specify the path to store files")
}
