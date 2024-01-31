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
	Keywords   []string   `yaml:"keywords"`
	Regex      string     `yaml:"regex"`
	Exclusions Exclusions `yaml:"exclusions"`
	Options    string     `yaml:"options"`
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

			excludedDirs := ""
			for _, dir := range config.Exclusions.Directories {
				excludedDirs += fmt.Sprintf(" --exclude-dir=%s", dir)
			}

			excludedFiles := ""
			for _, file := range config.Exclusions.Files {
				excludedFiles += fmt.Sprintf(" --exclude=%s", file)
			}

			for _, keyword := range config.Keywords {
				replacedRegex := strings.ReplaceAll(config.Regex, "{key}", keyword)
				out, err := exec.Command("bash", "-c", fmt.Sprintf("egrep %s '%s' ./ %s %s", config.Options, replacedRegex, excludedDirs, excludedFiles)).Output()
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

			if err := f.SaveAs("EgrepResults.xlsx"); err != nil {
				return err
			}
			return nil
		},
	}

	utCmd.AddCommand(egrepCmd)
	utCmd.Execute()
}
