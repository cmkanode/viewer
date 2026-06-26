package dbutils

import (
	"database/sql"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func AskForFolderAndImport(db *sql.DB, w fyne.Window, statusLabel *widget.Label, onDone func()) {
	statusLabel.SetText("No slides found. Please select a folder to import images.")
	ImportFolderDialog(db, w, statusLabel, onDone)
}
