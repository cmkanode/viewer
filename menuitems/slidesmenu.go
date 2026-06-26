package menuitems

import (
	"database/sql"
	"fmt"
	. "viewer/dbutils"
	. "viewer/models"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func SlideMenuItems(db *sql.DB, w fyne.Window, statusLabel *widget.Label, loadRandom func(), currentSlide **Slide) *fyne.Menu {
	randomShortcut := &desktop.CustomShortcut{
		KeyName:  fyne.KeyRight,
		Modifier: fyne.KeyModifierControl, // | fyne.KeyModifierShift,
	}
	randomMenuItem := fyne.NewMenuItem("Next Random", func() {
		loadRandom()
	})
	randomMenuItem.Shortcut = randomShortcut

	completedMenuItem := fyne.NewMenuItem("Mark as Completed", func() {
		if *currentSlide == nil {
			dialog.ShowError(fmt.Errorf("no slide loaded"), w)
			return
		}
		dialog.ShowConfirm("Confirm Mark as Completed", fmt.Sprintf("Mark image %d (%s) as completed?", (*currentSlide).ID, (*currentSlide).Name), func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := MarkImageSlideAsCompleted(db, (*currentSlide).ID); err != nil {
				dialog.ShowError(err, w)
				return
			}
			statusLabel.SetText("Image marked as completed. Loading next...")
			*currentSlide = nil
			loadRandom()
		}, w)
	})

	deleteMenuItem := fyne.NewMenuItem("Delete Image", func() {
		if *currentSlide == nil {
			dialog.ShowError(fmt.Errorf("no slide loaded"), w)
			return
		}
		dialog.ShowConfirm("Confirm Delete", fmt.Sprintf("Delete image %d (%s)?", (*currentSlide).ID, (*currentSlide).Name), func(confirmed bool) {
			if !confirmed {
				return
			}
			if err := DeleteImageSlide(db, (*currentSlide).ID); err != nil {
				dialog.ShowError(err, w)
				return
			}
		}, w)
	})

	slidesMenu := fyne.NewMenu("Slides",
		randomMenuItem,
		completedMenuItem,
		deleteMenuItem,
	)
	return slidesMenu
}
