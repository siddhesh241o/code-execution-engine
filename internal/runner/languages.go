package runner

import "fmt"

type LanguageConfig struct {
	Image      string
	Command    []string
	SourceFile string
}

var supportedLanguages = map[string]LanguageConfig{
	"python": {
		Image:      "engine-python",
		SourceFile: "solution.py",
		Command: []string{
			"sh", "-c",
			"/usr/bin/time -v -o metrics.txt python3 -u solution.py",
		},
	},
	"cpp": {
		Image: "engine-cpp",
		Command: []string{
			"sh", "-c",
			"g++ main.cpp -o app && /usr/bin/time -v -o metrics.txt ./app",
		},
		SourceFile: "main.cpp",
	},
	"go": {
		Image: "engine-go",
		Command: []string{
			"sh", "-c",
			"go build -o app main.go && /usr/bin/time -v -o metrics.txt ./app",
		},
		SourceFile: "main.go",
	},
	"java": {
		Image: "engine-java",
		Command: []string{
			"sh", "-c",
			"javac Main.java && /usr/bin/time -v -o metrics.txt java -Xmx100M -cp . Main",
		},
		SourceFile: "Main.java",
	},
	"javascript": {
		Image: "engine-node",
		Command: []string{
			"sh", "-c",
			"/usr/bin/time -v -o metrics.txt node solution.js",
		},
		SourceFile: "solution.js",
	},
}

func GetLanguageConfig(language string) (LanguageConfig, error) {
	config, exists := supportedLanguages[language]
	if !exists {
		return LanguageConfig{}, fmt.Errorf("language %s is not supported", language)
	}
	return config, nil
}
