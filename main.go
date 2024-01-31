package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	var utCmd = &cobra.Command{
		Use:   "ut",
		Short: "UT command",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ut")
		},
	}

	var egrepCmd = &cobra.Command{
		Use:                "egrep",
		Short:              "egrep command",
		DisableFlagParsing: true,
		Run: func(cmd *cobra.Command, args []string) {
			out, err := exec.Command("egrep", args...).CombinedOutput()
			if err != nil {
				fmt.Println(fmt.Sprintf("Error occurred: %v", err))
				return
			}
			fmt.Println(string(out))
		},
	}

	utCmd.AddCommand(egrepCmd)

	err := utCmd.Execute()
	if err != nil {
		fmt.Println(fmt.Sprintf("Error occurred: %v", strings.TrimSpace(err.Error())))
		return
	}
}
