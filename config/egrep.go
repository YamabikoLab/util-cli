package config

type Egrep struct {
	Keywords         []string `yaml:"keywords"`
	ConcurrencyLimit int      `yaml:"concurrencyLimit"`
	Options          string   `yaml:"options"`
	Regex            string   `yaml:"regex"`
	Exclusions       struct {
		Directories []string `yaml:"directories"`
		Files       []string `yaml:"files"`
	} `yaml:"exclusions"`
	TargetDir string `yaml:"targetDir"`
	Output    struct {
		Excel struct {
			Enable   bool   `yaml:"enable"`
			FilePath string `yaml:"filePath"`
			Sheet    struct {
				NameLimit int `yaml:"nameLimit"`
			} `yaml:"sheet"`
		} `yaml:"excel"`
		Text struct {
			Enable  bool   `yaml:"enable"`
			DirPath string `yaml:"dirPath"`
		}
	} `yaml:"output"`
}
