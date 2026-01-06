package runner

import (
	"testing"
)

func TestGetLanguageConfig(t *testing.T) {
	t.Run("Supported Language Python", func(t *testing.T) {
		_, err := GetLanguageConfig("python")
		if err != nil {
			t.Errorf("Unexpected error :%v", err)
		}
	})
}
