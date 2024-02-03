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
	"util-cli/config"
)

// ExcelSheetNameLimit Excelのシート名の最大文字数
const ExcelSheetNameLimit = 31

// InvalidExcelCharacters Excelのシート名に使えない文字
var InvalidExcelCharacters = map[rune]bool{
	'*':  true,
	'/':  true,
	'[':  true,
	']':  true,
	'?':  true,
	'\\': true,
	':':  true,
}

// loadConfig 設定ファイルを読み込む
func loadConfig(file string) (*config.Egrep, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	c := &config.Config{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling yaml: %w", err)
	}

	return &c.Egrep, nil
}

// Result コマンドの実行結果
type Result struct {
	Keyword string
	Error   error
}

// outputToTextFile テキストファイルに出力する
func outputToTextFile(keyword string, output string, dirPath string) error {
	// ファイルを作成する
	file, err := os.Create(filepath.Join(dirPath, fmt.Sprintf("%s.txt", keyword)))
	if err != nil {
		return err
	}

	// 関数終了時にファイルを閉じる
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	// 結果を書き込む
	_, err = file.WriteString(output)
	if err != nil {
		return err
	}

	return nil
}

// normalizeKeyword キーワードを正規化する
func normalizeKeyword(keyword string, limit int) string {
	normalized := make([]rune, 0, len(keyword))

	for _, ch := range keyword {
		if InvalidExcelCharacters[ch] {
			normalized = append(normalized, '_')
		} else {
			normalized = append(normalized, ch)
		}
	}

	result := string(normalized)
	if len(result) > limit {
		result = result[:limit]
	}

	return result
}

func createCommand(keyword, options, regex, targetDir, excludedDirs, excludedFiles string) string {
	replacedRegex := strings.ReplaceAll(regex, "{key}", keyword)
	return fmt.Sprintf("egrep %s '%s' %s %s %s", options, replacedRegex, targetDir, excludedDirs, excludedFiles)
}

// worker キーワードごとにegrepコマンドを実行し、結果をExcelファイル、または、テキストファイルに出力する
func worker(wg *sync.WaitGroup, index int, egrep *config.Egrep, keyword string, targetDir string, excludedDirs string, excludedFiles string, f *excelize.File, sheetNameLimit int, resultChan chan<- Result) {
	defer wg.Done()

	cmd := createCommand(keyword, egrep.Options, egrep.Regex, targetDir, excludedDirs, excludedFiles)
	out, err := exec.Command("bash", "-c", cmd).Output()

	var result Result
	result.Keyword = keyword
	result.Error = err

	if err == nil {
		if egrep.Output.Excel.Enable {
			sheetName := normalizeKeyword(keyword, sheetNameLimit)

			lines := strings.Split(string(out), "\n")
			for j := 0; j < len(lines); j++ {
				err := f.SetCellValue(sheetName, fmt.Sprintf("A%d", j+1), lines[j])
				if err != nil {
					return
				}
			}

			err = f.SetCellValue("result", fmt.Sprintf("A%d", index+1), keyword)
			if err != nil {
				return
			}
			err = f.SetCellValue("result", fmt.Sprintf("B%d", index+1), cmd)
			if err != nil {
				return
			}
		}

		if egrep.Output.Text.Enable {
			err = outputToTextFile(keyword, string(out), egrep.Output.Text.DirPath)
			if err != nil {
				result.Error = fmt.Errorf("writing to text file for keyword %s: %v", keyword, err)
			}
		}
	} else {
		result.Error = fmt.Errorf("keyword %s: %v", keyword, err)
	}

	resultChan <- result
}

// initExcelFile Excelファイルを初期化する
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

func listToStrings(items []string, format, separator string) string {
	str := ""
	for _, item := range items {
		str += fmt.Sprintf(format, item) + separator
	}
	return str
}

// RunEgrep egrepコマンドを実行する
func RunEgrep(_ *cobra.Command, _ []string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configFile := filepath.Join(homeDir, ".util-cli", "config.yml")
	egrep, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	excludedDirs := listToStrings(egrep.Exclusions.Directories, "--exclude-dir=%s", " ")
	excludedFiles := listToStrings(egrep.Exclusions.Files, "--exclude=%s", " ")
	targetDir := "."
	if egrep.TargetDir != "" {
		targetDir = egrep.TargetDir
	}

	// Excelとテキストの出力が両方無効の場合はメッセージを表示して終了
	if !egrep.Output.Excel.Enable && !egrep.Output.Text.Enable {
		fmt.Println("Excel and text outputs are both disabled. Nothing to do.")
		return nil
	}

	var f *excelize.File
	var sheetNameLimit int
	if egrep.Output.Excel.Enable {
		f, err = initExcelFile()
		if err != nil {
			return err
		}

		sheetNameLimit = ExcelSheetNameLimit
		if egrep.Output.Excel.Sheet.NameLimit != 0 {
			sheetNameLimit = egrep.Output.Excel.Sheet.NameLimit
		}

		for _, keyword := range egrep.Keywords {
			sheetName := normalizeKeyword(keyword, sheetNameLimit)
			if _, err := f.NewSheet(sheetName); err != nil {
				_, err := fmt.Fprintln(os.Stderr, err.Error())
				if err != nil {
					return err
				}
			}
		}
	}

	if egrep.Output.Text.Enable {
		err = os.MkdirAll(egrep.Output.Text.DirPath, 0755)
		if err != nil {
			return err
		}
	}
	concurrencyLimit := egrep.ConcurrencyLimit

	output := "EgrepResults.xlsx"
	if egrep.Output.Excel.FilePath != "" {
		output = egrep.Output.Excel.FilePath
	}

	var wg sync.WaitGroup
	sem := make(chan bool, concurrencyLimit)
	resultChan := make(chan Result, len(egrep.Keywords))

	for i, keyword := range egrep.Keywords {
		wg.Add(1)
		sem <- true // will block if there is already concurrencyLimit workers active
		go func(i int, keyword string) {
			worker(&wg, i+1, egrep, keyword, targetDir, excludedDirs, excludedFiles, f, sheetNameLimit, resultChan)
			<-sem // will only run once a worker has finished
		}(i, keyword)
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

	if egrep.Output.Excel.Enable {
		if err := f.SaveAs(output); err != nil {
			return err
		}
		fmt.Println("Excel output saved to:", output)
	}

	if egrep.Output.Text.Enable {
		fmt.Println("Text output saved to:", egrep.Output.Text.DirPath)
	}

	return nil
}
