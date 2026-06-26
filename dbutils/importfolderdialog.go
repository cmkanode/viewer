package dbutils

import (
	"database/sql"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ImportFolderDialog(db *sql.DB, w fyne.Window, statusLabel *widget.Label, onDone func()) {
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if uri == nil {
			statusLabel.SetText("No folder selected.")
			return
		}

		folder := uri.Path()
		statusLabel.SetText("Importing images...")
		if err := ImportImagesFromFolder(db, folder); err != nil {
			dialog.ShowError(err, w)
			return
		}
		statusLabel.SetText("Folder import complete.")
		onDone()
	}, w)
}
