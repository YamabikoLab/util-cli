package main

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			out, err := exec.Command("egrep", args...).CombinedOutput()
			if err != nil {
				return err
			}

			f := excelize.NewFile()

			lines := strings.Split(string(out), "\n")
			for i, line := range lines {
				// Excel file indexes start from 1, not 0
				f.SetCellValue("Sheet1", "A"+strconv.Itoa(i+1), line)
			}

			if err := f.SaveAs("EgrepResults.xlsx"); err != nil {
				return err
			}

			return nil
		},
	}

	utCmd.AddCommand(egrepCmd)
	utCmd.Execute()
}
