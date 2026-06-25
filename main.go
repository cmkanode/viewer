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
	Tags      []string
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
	tagsLabel := widget.NewLabel("Tags: ")
	image := canvas.NewImageFromResource(fyne.NewStaticResource("blank", []byte{}))
	image.FillMode = canvas.ImageFillContain
	image.SetMinSize(fyne.NewSize(1, 1))

	// Variable to hold the current slide
	var currentSlide *Slide

	loadRandom := func() {
		slide, err := selectRandomSlide(db)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		currentSlide = slide
		image.Resource = fyne.NewStaticResource(fmt.Sprintf("slide-%d", slide.ID), slide.Image)
		image.Refresh()
		statusLabel.SetText(fmt.Sprintf("%d: %s (%s)", slide.ID, slide.Name, completedText(slide.Completed)))

		// Update tags display
		if len(slide.Tags) > 0 {
			tagsLabel.SetText("Tags: " + fmt.Sprint(slide.Tags))
		} else {
			tagsLabel.SetText("Tags: (none)")
		}

		w.SetTitle(fmt.Sprintf("Image Viewer — %s", slide.Name))
	}

	nextButton := widget.NewButton("Next Random", func() {

		loadRandom()
	})

	addTagButton := widget.NewButton("Add Tag", func() {
		if currentSlide == nil {
			dialog.ShowError(fmt.Errorf("no slide loaded"), w)
			return
		}
		dialog.ShowEntryDialog("Add Tag", "Enter tag name:", func(value string) {
			if value == "" {
				return
			}
			if err := addTagToSlide(db, currentSlide.ID, value); err != nil {
				dialog.ShowError(err, w)
				return
			}
			// Reload tags
			tags, err := getTagsForSlide(db, currentSlide.ID)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			currentSlide.Tags = tags
			if len(tags) > 0 {
				tagsLabel.SetText("Tags: " + fmt.Sprint(tags))
			} else {
				tagsLabel.SetText("Tags: (none)")
			}
		}, w)
	})

	imageContainer := container.New(layout.NewStackLayout(), image)
	content := container.NewBorder(
		container.NewVBox(statusLabel, tagsLabel, container.NewHBox(nextButton, addTagButton)),
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
			importSingleImage(db, w, statusLabel, loadRandom)
		}),
		fyne.NewMenuItem("Import Folder", func() {
			importFolderDialog(db, w, statusLabel, loadRandom)
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() {
			app.Quit()
		}),
	)
	tagsMenu := fyne.NewMenu("Tags",
		fyne.NewMenuItem("Add Tag", func() {
			if currentSlide == nil {
				dialog.ShowError(fmt.Errorf("no slide loaded"), w)
				return
			}
			dialog.ShowEntryDialog("Add Tag", "Enter tag name:", func(value string) {
				if value == "" {
					return
				}
				if err := addTagToSlide(db, currentSlide.ID, value); err != nil {
					dialog.ShowError(err, w)
					return
				}
				// Reload tags
				tags, err := getTagsForSlide(db, currentSlide.ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				currentSlide.Tags = tags
				if len(tags) > 0 {
					tagsLabel.SetText("Tags: " + fmt.Sprint(tags))
				} else {
					tagsLabel.SetText("Tags: (none)")
				}
			}, w)
		}),
		fyne.NewMenuItem("Remove Tag", func() {
			if currentSlide == nil {
				dialog.ShowError(fmt.Errorf("no slide loaded"), w)
				return
			}
			if len(currentSlide.Tags) == 0 {
				dialog.ShowError(fmt.Errorf("no tags to remove"), w)
				return
			}
			dialog.ShowEntryDialog("Remove Tag", "Enter tag name to remove:", func(value string) {
				if value == "" {
					return
				}
				if err := removeTagFromSlide(db, currentSlide.ID, value); err != nil {
					dialog.ShowError(err, w)
					return
				}
				// Reload tags
				tags, err := getTagsForSlide(db, currentSlide.ID)
				if err != nil {
					dialog.ShowError(err, w)
					return
				}
				currentSlide.Tags = tags
				if len(tags) > 0 {
					tagsLabel.SetText("Tags: " + fmt.Sprint(tags))
				} else {
					tagsLabel.SetText("Tags: (none)")
				}
			}, w)
		}),
	)
	w.SetMainMenu(fyne.NewMainMenu(fileMenu, tagsMenu))

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

func importSingleImage(db *sql.DB, w fyne.Window, statusLabel *widget.Label, onDone func()) {
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

func importFolderDialog(db *sql.DB, w fyne.Window, statusLabel *widget.Label, onDone func()) {
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
		if err := importImagesFromFolder(db, folder); err != nil {
			dialog.ShowError(err, w)
			return
		}
		statusLabel.SetText("Folder import complete.")
		onDone()
	}, w)
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
	importFolderDialog(db, w, statusLabel, onDone)
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
	tags, err := getTagsForSlide(db, slide.ID)
	if err != nil {
		return nil, err
	}
	slide.Tags = tags
	return &slide, nil
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS slides (
		id INTEGER PRIMARY KEY,
		name TEXT,
		image BLOB,
		completed BOOLEAN NOT NULL DEFAULT FALSE
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL UNIQUE
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS slide_tags (
		id INTEGER PRIMARY KEY,
		slide_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		FOREIGN KEY (slide_id) REFERENCES slides(id) ON DELETE CASCADE,
		FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE,
		UNIQUE(slide_id, tag_id)
	)`)
	return err
}

func getTagsForSlide(db *sql.DB, slideID int) ([]string, error) {
	rows, err := db.Query(`SELECT t.name FROM tags t
		INNER JOIN slide_tags st ON t.id = st.tag_id
		WHERE st.slide_id = ?
		ORDER BY t.name`, slideID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func addTagToSlide(db *sql.DB, slideID int, tagName string) error {
	if tagName == "" {
		return fmt.Errorf("tag name cannot be empty")
	}
	// Insert tag if it doesn't exist
	_, err := db.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tagName)
	if err != nil {
		return err
	}

	// Get the tag ID
	var tagID int
	err = db.QueryRow("SELECT id FROM tags WHERE name = ?", tagName).Scan(&tagID)
	if err != nil {
		return err
	}

	// Insert the relationship
	_, err = db.Exec("INSERT OR IGNORE INTO slide_tags (slide_id, tag_id) VALUES (?, ?)", slideID, tagID)
	return err
}

func removeTagFromSlide(db *sql.DB, slideID int, tagName string) error {
	_, err := db.Exec(`DELETE FROM slide_tags 
		WHERE slide_id = ? AND tag_id = (SELECT id FROM tags WHERE name = ?)`, slideID, tagName)
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
