package dbutils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ImportSingleImage(db *sql.DB, w fyne.Window, statusLabel *widget.Label, onDone func()) {
	dialog.ShowFileOpen(func(uri fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if uri == nil {
			statusLabel.SetText("No file selected.")
			return
		}
		defer uri.Close()

		// Check if file is a valid image extension
		extensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true}
		ext := filepath.Ext(uri.URI().String())
		if !extensions[ext] {
			dialog.ShowError(fmt.Errorf("unsupported file format: %s", ext), w)
			return
		}

		// Read file data
		data, err := os.ReadFile(uri.URI().Path())
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		// Insert into database
		_, err = db.Exec("INSERT INTO slides (name, image, completed) VALUES (?, ?, FALSE)", filepath.Base(uri.URI().Path()), data)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		statusLabel.SetText("Image imported successfully.")
		onDone()
	}, w)
}
