package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

const ExcelSheetNameLimit = 31

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
			homeDir, err := os.UserHomeDir()
			if err != nil {
				_, err := fmt.Fprintln(os.Stderr, err)
				if err != nil {
					return err
				}
				os.Exit(1)
			}

			configFile := filepath.Join(homeDir, ".util-cli", "config.yml")
			data, err := os.ReadFile(configFile)
			if err != nil {
				_, err := fmt.Fprintln(os.Stderr, err)
				if err != nil {
					return err
				}
				os.Exit(1)
			}

			config := Config{}

			err = yaml.Unmarshal(data, &config)
			if err != nil {
				_, err := fmt.Fprintln(os.Stderr, err)
				if err != nil {
					return err
				}
				os.Exit(1)
			}

			f := excelize.NewFile()

			// Add a new sheet named "result"
			_, err = f.NewSheet("result")
			if err != nil {
				return err
			}

			// Set the headers for the "result" sheet
			err = f.SetCellValue("result", "A1", "keyword")
			if err != nil {
				return err
			}
			err = f.SetCellValue("result", "B1", "cmd")
			if err != nil {
				return err
			}

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

			var noResultKeywords []string

			for i, keyword := range egrepConfig.Keywords {
				replacedRegex := strings.ReplaceAll(egrepConfig.Regex, "{key}", keyword)
				cmd := fmt.Sprintf("egrep %s '%s' %s %s %s", egrepConfig.Options, replacedRegex, targetDir, excludedDirs, excludedFiles)
				out, err := exec.Command("bash", "-c", cmd).Output()

				// Output the keyword and command to the "result" sheet
				err = f.SetCellValue("result", fmt.Sprintf("A%d", i+2), keyword)
				if err != nil {
					return err
				}
				err = f.SetCellValue("result", fmt.Sprintf("B%d", i+2), cmd)
				if err != nil {
					return err
				}

				if err != nil {
					if len(strings.TrimSpace(string(out))) == 0 {
						noResultKeywords = append(noResultKeywords, keyword)
					}
					_, err := fmt.Fprintln(os.Stderr, err.Error())
					if err != nil {
						return err
					}
					continue
				}

				sheetName := keyword

				if len(sheetName) > ExcelSheetNameLimit {
					sheetName = sheetName[:ExcelSheetNameLimit] // Truncate and prepend with index to ensure uniqueness
				}

				_, err = f.NewSheet(sheetName)
				if err != nil {
					_, err := fmt.Fprintln(os.Stderr, err.Error())
					if err != nil {
						return err
					}
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
	err := utCmd.Execute()
	if err != nil {
		return
	}
}
