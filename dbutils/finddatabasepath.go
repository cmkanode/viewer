package dbutils

import (
	"os"
	"path/filepath"
)

func FindDatabasePath(workDir, exeDir string) (string, error) {
	candidates := []string{
		filepath.Join(workDir, "viewer.db"),
		filepath.Join(exeDir, "viewer.db"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return filepath.Join(workDir, "viewer.db"), nil
}
