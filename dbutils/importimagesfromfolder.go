package dbutils

import (
	"database/sql"
	"os"
	"path/filepath"
)

func ImportImagesFromFolder(db *sql.DB, folder string) error {
	extensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true}
	return filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !extensions[filepath.Ext(path)] {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = db.Exec("INSERT INTO slides (name, image, completed) VALUES (?, ?, FALSE)", filepath.Base(path), data)
		return err
	})
}
