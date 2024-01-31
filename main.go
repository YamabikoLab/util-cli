package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Egrep struct {
		Keywords   []string `yaml:"keywords"`
		Regex      string   `yaml:"regex"`
		Options    string   `yaml:"options"`
		Exclusions struct {
			Directories []string `yaml:"directories"`
			Files       []string `yaml:"files"`
		} `yaml:"exclusions"`
		TargetDir string `yaml:"targetDir"`
		Output    string `yaml:"output"`
	} `yaml:"egrep"`
}

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
			config := Config{}

			data, err := os.ReadFile("config.yml")
			if err != nil {
				return err
			}

			err = yaml.UnmarshalStrict(data, &config)
			if err != nil {
				return err
			}

			f := excelize.NewFile()

			// Add a new sheet named "result"
			f.NewSheet("result")

			// Set the headers for the "result" sheet
			f.SetCellValue("result", "A1", "keyword")
			f.SetCellValue("result", "B1", "cmd")

			egrepConfig := config.Egrep

			excludedDirs := ""
			for _, dir := range egrepConfig.Exclusions.Directories {
				excludedDirs += fmt.Sprintf(" --exclude-dir=%s", dir)
			}

			excludedFiles := ""
			for _, file := range egrepConfig.Exclusions.Files {
				excludedFiles += fmt.Sprintf(" --exclude=%s", file)
			}

			targetDir := "."
			if egrepConfig.TargetDir != "" {
				targetDir = egrepConfig.TargetDir
			}

			output := "EgrepResults.xlsx"
			if egrepConfig.Output != "" {
				output = egrepConfig.Output
			}

			noResultKeywords := []string{}

			for i, keyword := range egrepConfig.Keywords {
				replacedRegex := strings.ReplaceAll(egrepConfig.Regex, "{key}", keyword)
				cmd := fmt.Sprintf("egrep %s '%s' %s %s %s", egrepConfig.Options, replacedRegex, targetDir, excludedDirs, excludedFiles)
				out, err := exec.Command("bash", "-c", cmd).Output()

				// Output the keyword and command to the "result" sheet
				f.SetCellValue("result", fmt.Sprintf("A%d", i+2), keyword)
				f.SetCellValue("result", fmt.Sprintf("B%d", i+2), cmd)

				if err != nil {
					if len(strings.TrimSpace(string(out))) == 0 {
						noResultKeywords = append(noResultKeywords, keyword)
					}
					fmt.Fprintln(os.Stderr, err.Error())
					continue
				}

				_, err = f.NewSheet(keyword)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					continue
				}

				lines := strings.Split(string(out), "\n")
				for j := 1; j <= len(lines); j++ {
					_ = f.SetCellValue(keyword, fmt.Sprintf("A%d", j), lines[j-1])
				}
			}

			if err := f.SaveAs(output); err != nil {
				return err
			}

			if len(noResultKeywords) > 0 {
				fmt.Println("No results keywords:")
				for _, keyword := range noResultKeywords {
					fmt.Println(keyword)
				}
			}
			fmt.Println("Output saved to:", output)

			return nil
		},
	}

	utCmd.AddCommand(egrepCmd)
	utCmd.Execute()
}
