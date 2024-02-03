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
	"sync"
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
		Output    struct {
			Excel struct {
				FilePath string `yaml:"filePath"`
				Sheet    struct {
					NameLimit int `yaml:"nameLimit"`
				} `yaml:"sheet"`
			} `yaml:"excel"`
		} `yaml:"output"`
	} `yaml:"egrep"`
}

const ExcelSheetNameLimit = 31

func loadConfig(file string) (*Config, error) {
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

func createCommand(keyword, options, regex, targetDir, excludedDirs, excludedFiles string) string {
	replacedRegex := strings.ReplaceAll(regex, "{key}", keyword)
	return fmt.Sprintf("egrep %s '%s' %s %s %s", options, replacedRegex, targetDir, excludedDirs, excludedFiles)
}

func listToStrings(items []string, format, separator string) string {
	str := ""
	for _, item := range items {
		str += fmt.Sprintf(format, item) + separator
	}
	return str
}

type Result struct {
	Keyword string
	Error   error
}

func worker(wg *sync.WaitGroup, f *excelize.File, index int, keyword string, config *Config, targetDir string, excludedDirs string, excludedFiles string, sheetNameLimit int, resultChan chan<- Result) {
	defer wg.Done()

	cmd := createCommand(keyword, config.Egrep.Options, config.Egrep.Regex, targetDir, excludedDirs, excludedFiles)
	out, err := exec.Command("bash", "-c", cmd).Output()

	var result Result
	result.Keyword = keyword
	result.Error = err

	if err == nil {
		// success case: output the result to a new sheet
		sheetName := keyword
		if len(sheetName) > sheetNameLimit {
			sheetName = sheetName[:sheetNameLimit]
		}

		_, err := f.NewSheet(sheetName)
		if err != nil {
			_, err := fmt.Fprintln(os.Stderr, err.Error())
			if err != nil {
				return
			}
		} else {
			lines := strings.Split(string(out), "\n")
			for j := 0; j < len(lines); j++ {
				_ = f.SetCellValue(sheetName, fmt.Sprintf("A%d", j+1), lines[j])
			}
		}
	} else {
		// failed case: put the keyword and error message to the result
		result.Error = fmt.Errorf("keyword %s: %v", keyword, err)
	}

	// Common: Output the keyword and command to the "result" sheet
	_ = f.SetCellValue("result", fmt.Sprintf("A%d", index+1), keyword)
	_ = f.SetCellValue("result", fmt.Sprintf("B%d", index+1), cmd)

	resultChan <- result
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
	config, err := loadConfig(configFile)
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

	excludedDirs := listToStrings(egrepConfig.Exclusions.Directories, "--exclude-dir=%s", " ")

	excludedFiles := listToStrings(egrepConfig.Exclusions.Files, "--exclude=%s", " ")

	targetDir := "."
	if egrepConfig.TargetDir != "" {
		targetDir = egrepConfig.TargetDir
	}

	output := "EgrepResults.xlsx"
	if egrepConfig.Output.Excel.FilePath != "" {
		output = egrepConfig.Output.Excel.FilePath
	}

	sheetNameLimit := ExcelSheetNameLimit
	if egrepConfig.Output.Excel.Sheet.NameLimit != 0 {
		sheetNameLimit = egrepConfig.Output.Excel.Sheet.NameLimit
	}

	var noResultKeywords []string

	var wg sync.WaitGroup
	resultChan := make(chan Result, len(egrepConfig.Keywords))

	for i, keyword := range egrepConfig.Keywords {
		wg.Add(1)
		go worker(&wg, f, i+1, keyword, config, targetDir, excludedDirs, excludedFiles, sheetNameLimit, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	noResultKeywords = make([]string, 0)
	for result := range resultChan {
		if result.Error != nil {
			// error occurred in one of the worker, you may want to append it to the noResultKeywords
			noResultKeywords = append(noResultKeywords, result.Keyword)
		}
	}

	if err := f.SaveAs(output); err != nil {
		return err
	}

	fmt.Println("Output saved to:", output)

	return nil
}
