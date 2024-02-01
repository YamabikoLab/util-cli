package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"util-cli/commands"
)

func main() {
	var utCmd = &cobra.Command{
		Use:   "ut",
		Short: "UT command",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}

	var egrepCmd = &cobra.Command{
		Use:                "egrep",
		Short:              "egrep command",
		DisableFlagParsing: true,
		RunE:               commands.RunEgrep,
	}

	utCmd.AddCommand(egrepCmd)
	if err := utCmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		return
	}
}
