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

type Exclusions struct {
	Directories []string `yaml:"directories"`
	Files       []string `yaml:"files"`
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

			for _, keyword := range egrepConfig.Keywords {
				replacedRegex := strings.ReplaceAll(egrepConfig.Regex, "{key}", keyword)
				out, err := exec.Command("bash", "-c", fmt.Sprintf("egrep %s '%s' %s %s %s", egrepConfig.Options, replacedRegex, targetDir, excludedDirs, excludedFiles)).Output()
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					continue
				}

				_, err = f.NewSheet(keyword)
				if err != nil {
					fmt.Fprintln(os.Stderr, err.Error())
					continue
				}

				lines := strings.Split(string(out), "\n")
				for i := 1; i <= len(lines); i++ {
					_ = f.SetCellValue(keyword, fmt.Sprintf("A%d", i), lines[i-1])
				}
			}

			if err := f.SaveAs(output); err != nil {
				return err
			}

			fmt.Println("Output saved to:", output)

			return nil
		},
	}

	utCmd.AddCommand(egrepCmd)
	utCmd.Execute()
}
