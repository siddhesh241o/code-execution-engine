package runner

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileManager struct {
}

func NewFileManager() *FileManager {
	return &FileManager{}
}

func (fm *FileManager) CreateWorkspace(code string, filename string) (string, func(), error) {
	workdir, err := os.MkdirTemp("", "code-box-*")
	if err != nil {
		return "", nil, fmt.Errorf("failed to created tmp directory: %v", err)
	}
	cleanup := func() {
		_ = os.RemoveAll(workdir)
	}

	filePath := filepath.Join(workdir, filename)
	if err = os.WriteFile(filePath, []byte(code), 0600); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to created code file: %v", err)
	}
	return workdir, cleanup, err
}
