package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/xuri/excelize/v2"
	"gopkg.in/yaml.v3"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func LoadConfig(file string) (*Config, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	c := &Config{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling yaml: %w", err)
	}

	return c, nil
}

func CreateCommand(keyword, options, regex, targetDir, excludedDirs, excludedFiles string) string {
	replacedRegex := strings.ReplaceAll(regex, "{key}", keyword)
	return fmt.Sprintf("egrep %s '%s' %s %s %s", options, replacedRegex, targetDir, excludedDirs, excludedFiles)
}

func ListToStrings(items []string, format, separator string) string {
	str := ""
	for _, item := range items {
		str += fmt.Sprintf(format, item) + separator
	}
	return str
}

func RunEgrep(_ *cobra.Command, _ []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return err
		}
		os.Exit(1)
	}

	configFile := filepath.Join(homeDir, ".util-cli", "config.yml")
	config, err := LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	f := excelize.NewFile()
	_, err = f.NewSheet("result")
	if err != nil {
		return fmt.Errorf("creating new sheet: %w", err)
	}

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

	excludedDirs := ListToStrings(egrepConfig.Exclusions.Directories, "--exclude-dir=%s", " ")

	excludedFiles := ListToStrings(egrepConfig.Exclusions.Files, "--exclude=%s", " ")

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
		cmd := CreateCommand(keyword, egrepConfig.Options, egrepConfig.Regex, targetDir, excludedDirs, excludedFiles)
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
}
