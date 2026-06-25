package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	_ "modernc.org/sqlite"
)

type Slide struct {
	ID        int
	Name      string
	Image     []byte
	Completed bool
}

// Windows API constants for SetThreadExecutionState
const (
	ES_CONTINUOUS       = 0x80000000
	ES_DISPLAY_REQUIRED = 0x00000002
	ES_SYSTEM_REQUIRED  = 0x00000001
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var setThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")

// preventSleep tells Windows not to sleep or turn off display while app is running
func preventSleep() {
	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS | ES_DISPLAY_REQUIRED | ES_SYSTEM_REQUIRED))
}

// allowSleep restores normal Windows sleep behavior
func allowSleep() {
	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS))
}

func main() {
	// Prevent system sleep and display timeout while app is running
	preventSleep()
	defer allowSleep()

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

	dbPath, err := findDatabasePath(workDir, exeDir)
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

	if err := ensureSchema(db); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create schema: %v\n", err)
		os.Exit(1)
	}

	app := app.New()
	w := app.NewWindow("Image Viewer")
	statusLabel := widget.NewLabel("Loading random slide...")
	image := canvas.NewImageFromResource(fyne.NewStaticResource("blank", []byte{}))
	image.FillMode = canvas.ImageFillContain
	image.SetMinSize(fyne.NewSize(1, 1))

	loadRandom := func() {
		slide, err := selectRandomSlide(db)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		image.Resource = fyne.NewStaticResource(fmt.Sprintf("slide-%d", slide.ID), slide.Image)
		image.Refresh()
		statusLabel.SetText(fmt.Sprintf("%d: %s (%s)", slide.ID, slide.Name, completedText(slide.Completed)))
		w.SetTitle(fmt.Sprintf("Image Viewer — %s", slide.Name))
	}

	nextButton := widget.NewButton("Next Random", func() {

		loadRandom()
	})

	imageContainer := container.New(layout.NewStackLayout(), image)
	content := container.NewBorder(
		container.NewVBox(statusLabel, nextButton),
		nil,
		nil,
		nil,
		imageContainer,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(800, 620))

	slideCount, err := countSlides(db)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to count slides: %v\n", err)
		os.Exit(1)
	}

	if slideCount == 0 {
		askForFolderAndImport(db, w, statusLabel, loadRandom)
	} else {
		loadRandom()
	}

	w.ShowAndRun()
}

func countSlides(db *sql.DB) (int, error) {
	row := db.QueryRow("SELECT COUNT(*) FROM slides")
	var count int
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func askForFolderAndImport(db *sql.DB, w fyne.Window, statusLabel *widget.Label, onDone func()) {
	statusLabel.SetText("No slides found. Please select a folder to import images.")
	dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if uri == nil {
			statusLabel.SetText("No folder selected. Import canceled.")
			return
		}

		folder := uri.Path()
		statusLabel.SetText("Importing images...")
		if err := importImagesFromFolder(db, folder); err != nil {
			dialog.ShowError(err, w)
			return
		}
		statusLabel.SetText("Import complete.")
		onDone()
	}, w)
}

func importImagesFromFolder(db *sql.DB, folder string) error {
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

func selectRandomSlide(db *sql.DB) (*Slide, error) {
	row := db.QueryRow("SELECT id, name, image, completed FROM slides WHERE completed = FALSE ORDER BY RANDOM() LIMIT 1")
	var slide Slide
	if err := row.Scan(&slide.ID, &slide.Name, &slide.Image, &slide.Completed); err != nil {
		return nil, err
	}
	return &slide, nil
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS slides (
		id INTEGER PRIMARY KEY,
		name TEXT,
		image BLOB,
		completed BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	return err
}

func completedText(done bool) string {
	if done {
		return "completed"
	}
	return "pending"
}

func findDatabasePath(workDir, exeDir string) (string, error) {
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
