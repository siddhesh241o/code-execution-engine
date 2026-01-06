package runner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateWorkspace(t *testing.T){
	tests := []struct{
		name string 
		code string 
		filename string
	}{
		{
			name: "Python Test",
			code: "print('hello')",
			filename: "main.py",
		},
		{
			name: "Empty File",
			code: "",
			filename: "main.py",
		},
	}
	fm := NewFileManager()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T){
			workdir, cleanup, err := fm.CreateWorkspace(test.code, test.filename)
			if err != nil {
				t.Fatalf("failed creating a workspace: %v", err)
			}
			defer cleanup()
			filepath := filepath.Join(workdir, test.filename)
			data, err := os.ReadFile(filepath)
			if err != nil {
				t.Fatalf("could not read file: %v", err)
			}
			if string(data) != test.code {
				t.Errorf("content mismatch: expected %v, got %v", test.code, string(data))
			}
		})
	}
}