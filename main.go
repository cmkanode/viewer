package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "image/gif"
	_ "image/jpeg"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	_ "modernc.org/sqlite"

	. "viewer/dbutils"
	. "viewer/imgutils"
	. "viewer/menuitems"
	. "viewer/models"
	. "viewer/winutils"
)

func main() {
	// Prevent system sleep and display timeout while app is running
	PreventSleep()
	defer AllowSleep()

	workDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to determine working directory: %v\n", err)
		os.Exit(1)
	}

	exeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to determine executable path: %v\n", err)
		os.Exit(1)
	}

	dbPath, err := FindDatabasePath(workDir, exeDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database %q: %v\n", dbPath, err)
		os.Exit(1)
	}
	defer db.Close()

	if err := EnsureSchema(db); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create schema: %v\n", err)
		os.Exit(1)
	}

	app := app.New()
	w := app.NewWindow("Image Viewer")
	statusLabel := widget.NewLabel("Loading random slide...")
	tagsLabel := widget.NewLabel("Tags: ")

	placeholder := MakePlaceholderPNG()
	image := canvas.NewImageFromResource(fyne.NewStaticResource("blank", placeholder))
	image.FillMode = canvas.ImageFillContain
	image.SetMinSize(fyne.NewSize(1, 1))

	// Variable to hold the current slide
	var currentSlide *Slide

	loadRandom := func() {
		slide, err := SelectRandomSlide(db)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		if !IsValidImage(slide.Image) {
			errMsg := fmt.Sprintf("Slide %d failed to decode: unknown or invalid image format", slide.ID)
			fmt.Fprintln(os.Stderr, errMsg)
			dialog.ShowError(fmt.Errorf("Slide %d failed to decode: unknown or invalid image format", slide.ID), w)
			return
		}

		currentSlide = slide
		image.Resource = fyne.NewStaticResource(fmt.Sprintf("slide-%d", slide.ID), slide.Image)
		image.Refresh()
		statusLabel.SetText(fmt.Sprintf("%d: %s", slide.ID, slide.Name))

		// Update tags display
		if len(slide.Tags) > 0 {
			tagsLabel.SetText("Tags: " + fmt.Sprint(slide.Tags))
		} else {
			tagsLabel.SetText("Tags: (none)")
		}

		w.SetTitle(fmt.Sprintf("Image Viewer — %d: %s", slide.ID, slide.Name))
	}

	imageContainer := container.New(layout.NewStackLayout(), image)
	content := container.NewBorder(
		container.NewVBox(tagsLabel),
		nil,
		nil,
		nil,
		imageContainer,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 620))

	// Create menu bar
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Import Image", func() {
			ImportSingleImage(db, w, statusLabel, loadRandom)
		}),
		fyne.NewMenuItem("Import Folder", func() {
			ImportFolderDialog(db, w, statusLabel, loadRandom)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			app.Quit()
		}),
	)

	slidesMenu := SlideMenuItems(db, w, statusLabel, loadRandom, &currentSlide)
	tagsMenu := TagsMenuItems(db, w, &currentSlide, tagsLabel)

	w.SetMainMenu(fyne.NewMainMenu(fileMenu, slidesMenu, tagsMenu))

	slideCount, err := CountSlides(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to count slides: %v\n", err)
		os.Exit(1)
	}

	if slideCount == 0 {
		AskForFolderAndImport(db, w, statusLabel, loadRandom)
	} else {
		loadRandom()
	}

	w.ShowAndRun()
}

func completedText(done bool) string {
	if done {
		return "completed"
	}
	return "pending"
}
