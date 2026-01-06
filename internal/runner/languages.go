package runner

import "fmt"

type LanguageConfig struct {
	Image       string
	Command     []string
	SourceFile  string
	MemoryLimit int64
}

var supportedLanguages = map[string]LanguageConfig{
	"python": {
		Image:       "python:3.12-alpine",
		Command:     []string{"python", "/code/main.py"},
		SourceFile:  "main.py",
		MemoryLimit: 0,
	},
}

func GetLanguageConfig(language string) (LanguageConfig, error) {
	config, exists := supportedLanguages[language]
	if !exists {
		return LanguageConfig{}, fmt.Errorf("language %s is not supported", language)
	}
	return config, nil
}
