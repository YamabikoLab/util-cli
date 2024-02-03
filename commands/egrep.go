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

type Result struct {
	Keyword string
	Error   error
}

func worker(wg *sync.WaitGroup, f *excelize.File, index int, keyword string, config *Config, sheetNameLimit int, resultChan chan<- Result) {
	defer wg.Done()

	createCommand := func(keyword, options, regex, targetDir, excludedDirs, excludedFiles string) string {
		replacedRegex := strings.ReplaceAll(regex, "{key}", keyword)
		return fmt.Sprintf("egrep %s '%s' %s %s %s", options, replacedRegex, targetDir, excludedDirs, excludedFiles)
	}

	listToStrings := func(items []string, format, separator string) string {
		str := ""
		for _, item := range items {
			str += fmt.Sprintf(format, item) + separator
		}
		return str
	}

	egrep := config.Egrep
	excludedDirs := listToStrings(egrep.Exclusions.Directories, "--exclude-dir=%s", " ")
	excludedFiles := listToStrings(egrep.Exclusions.Files, "--exclude=%s", " ")
	targetDir := "."
	if egrep.TargetDir != "" {
		targetDir = egrep.TargetDir
	}

	cmd := createCommand(keyword, egrep.Options, egrep.Regex, targetDir, excludedDirs, excludedFiles)
	out, err := exec.Command("bash", "-c", cmd).Output()

	var result Result
	result.Keyword = keyword
	result.Error = err

	if err == nil {
		sheetName := keyword
		if len(sheetName) > sheetNameLimit {
			sheetName = sheetName[:sheetNameLimit]
		}

		lines := strings.Split(string(out), "\n")
		for j := 0; j < len(lines); j++ {
			err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", j+1), lines[j])
			if err != nil {
				return
			}
		}
	} else {
		result.Error = fmt.Errorf("keyword %s: %v", keyword, err)
	}

	err = f.SetCellValue("result", fmt.Sprintf("A%d", index+1), keyword)
	if err != nil {
		return
	}
	err = f.SetCellValue("result", fmt.Sprintf("B%d", index+1), cmd)
	if err != nil {
		return
	}

	resultChan <- result
}

func initExcelFile() (*excelize.File, error) {
	f := excelize.NewFile()
	if _, err := f.NewSheet("result"); err != nil {
		return nil, err
	}

	if err := f.SetCellValue("result", "A1", "keyword"); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("result", "B1", "cmd"); err != nil {
		return nil, err
	}

	return f, nil
}

func RunEgrep(_ *cobra.Command, _ []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(homeDir, ".util-cli", "config.yml")
	config, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	f, err := initExcelFile()
	if err != nil {
		return err
	}

	egrep := config.Egrep
	output := "EgrepResults.xlsx"
	if egrep.Output.Excel.FilePath != "" {
		output = egrep.Output.Excel.FilePath
	}

	sheetNameLimit := ExcelSheetNameLimit
	if egrep.Output.Excel.Sheet.NameLimit != 0 {
		sheetNameLimit = egrep.Output.Excel.Sheet.NameLimit
	}

	var wg sync.WaitGroup
	resultChan := make(chan Result, len(egrep.Keywords))

	for _, keyword := range egrep.Keywords {
		sheetName := keyword
		if len(sheetName) > sheetNameLimit {
			sheetName = sheetName[:sheetNameLimit]
		}

		if _, err := f.NewSheet(sheetName); err != nil {
			_, err := fmt.Fprintln(os.Stderr, err.Error())
			if err != nil {
				return err
			}
		}
	}

	for i, keyword := range egrep.Keywords {
		wg.Add(1)
		go worker(&wg, f, i+1, keyword, config, sheetNameLimit, resultChan)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	noResultKeywords := make([]string, 0)
	for result := range resultChan {
		if result.Error != nil {
			noResultKeywords = append(noResultKeywords, result.Keyword)
		}
	}

	if err := f.SaveAs(output); err != nil {
		return err
	}

	fmt.Println("Output saved to:", output)

	return nil
}
