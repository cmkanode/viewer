package menuitems

import (
	"database/sql"
	"fmt"
	. "viewer/dbutils"
	. "viewer/models"

	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func TagsMenuItems(db *sql.DB, w fyne.Window, currentSlide **Slide, tagsLabel *widget.Label) *fyne.Menu {
	tagShortcut := &desktop.CustomShortcut{
		KeyName:  fyne.KeyT,
		Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift,
	}
	addTagMenuItem := fyne.NewMenuItem("Add Tag", func() {
		if *currentSlide == nil {
			dialog.ShowError(fmt.Errorf("no slide loaded"), w)
			return
		}

		entry := widget.NewEntry()
		entry.SetPlaceHolder("Enter tag name")

		var dlg dialog.Dialog
		entry.OnSubmitted = func(value string) {
			if strings.TrimSpace(value) == "" {
				return
			}
			if err := AddTagToSlide(db, (*currentSlide).ID, value); err != nil {
				dialog.ShowError(err, w)
				return
			}
			// Reload tags
			tags, err := GetTagsForSlide(db, (*currentSlide).ID)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			(*currentSlide).Tags = tags
			if len(tags) > 0 {
				tagsLabel.SetText("Tags: " + fmt.Sprint(tags))
			} else {
				tagsLabel.SetText("Tags: (none)")
			}
			if dlg != nil {
				dlg.Hide()
			}
		}

		form := widget.NewForm(
			widget.NewFormItem("Tag", entry),
		)

		dlg = dialog.NewCustom("Add Tag", "Cancel", container.NewVBox(form), w)
		dlg.Resize(fyne.NewSize(300, 200))
		dlg.Show()
		w.Canvas().Focus(entry)
	})
	addTagMenuItem.Shortcut = tagShortcut

	return fyne.NewMenu("Tags",
		addTagMenuItem,
		fyne.NewMenuItem("Remove Tag", func() {
			if *currentSlide == nil {
				dialog.ShowError(fmt.Errorf("no slide loaded"), w)
				return
			}
			if len((*currentSlide).Tags) == 0 {
				dialog.ShowError(fmt.Errorf("no tags to remove"), w)
				return
			}
			dialog.ShowEntryDialog("Remove Tag", "Enter tag name to remove:", func(value string) {
				if value == "" {
					return
				}
				if err := RemoveTagFromSlide(db, (*currentSlide).ID, value); err != nil {
					dialog.ShowError(err, w)
					return
				}
				// Reload tags
				tags, err := GetTagsForSlide(db, (*currentSlide).ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				(*currentSlide).Tags = tags
				if len(tags) > 0 {
					tagsLabel.SetText("Tags: " + fmt.Sprint(tags))
				} else {
					tagsLabel.SetText("Tags: (none)")
				}
			}, w)
		}),
	)
}
