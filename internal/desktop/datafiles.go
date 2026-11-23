package desktop

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vanam-san/fittrack/internal/database"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type DataFilesService struct {
	files *database.DataFiles
	path  string
}

func NewDataFilesService(files *database.DataFiles, path string) *DataFilesService {
	return &DataFilesService{files: files, path: path}
}

func (s *DataFilesService) Export(ctx context.Context, format string) (string, error) {
	if format != "csv" && format != "json" {
		return "", fmt.Errorf("choose CSV or JSON")
	}
	path, err := s.saveDialog("", "fittrack-records."+format, strings.ToUpper(format)+" files", format)
	if err != nil || path == "" {
		return "", err
	}
	return path, s.files.Export(ctx, path, format)
}

func (s *DataFilesService) Backup(ctx context.Context) (string, error) {
	dir := database.BackupDirectory(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("open backup folder: %w", err)
	}
	path, err := s.saveDialog(dir, "fittrack-"+time.Now().Format("20060102-150405")+".sqlite3", "FitTrack backups", "sqlite3")
	if err != nil || path == "" {
		return "", err
	}
	return path, s.files.Backup(ctx, path)
}

func (s *DataFilesService) ChooseInput(operation string) (string, error) {
	app := application.Get()
	dialog := app.Dialog.OpenFile().AttachToWindow(app.Window.Current()).CanChooseFiles(true)
	switch operation {
	case "import":
		dialog.SetTitle("Choose a FitTrack export").AddFilter("FitTrack exports", "*.csv;*.json")
	case "restore":
		dialog.SetTitle("Choose a FitTrack backup").AddFilter("FitTrack backups", "*.sqlite3")
		if dir := database.BackupDirectory(s.path); isDirectory(dir) {
			dialog.SetDirectory(dir)
		}
	default:
		return "", fmt.Errorf("unknown file operation")
	}
	return dialog.PromptForSingleSelection()
}

func (s *DataFilesService) Replace(ctx context.Context, operation, path string) (string, error) {
	switch operation {
	case "import":
		return s.files.Import(ctx, path)
	case "restore":
		checkpoint, err := s.files.Restore(ctx, path)
		if err != nil && checkpoint != "" {
			return "", fmt.Errorf("restore failed: %w. Recovery backup saved to %s", err, checkpoint)
		}
		return checkpoint, err
	default:
		return "", fmt.Errorf("unknown file operation")
	}
}

func (s *DataFilesService) saveDialog(dir, filename, label, extension string) (string, error) {
	app := application.Get()
	dialog := app.Dialog.SaveFile().AttachToWindow(app.Window.Current()).
		SetFilename(filename).CanCreateDirectories(true).AddFilter(label, "*."+extension)
	if dir != "" {
		dialog.SetDirectory(dir)
	}
	path, err := dialog.PromptForSingleSelection()
	if err != nil || path == "" {
		return path, err
	}
	if !strings.EqualFold(filepath.Ext(path), "."+extension) {
		return "", fmt.Errorf("save the file with the .%s extension", extension)
	}
	return path, nil
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
