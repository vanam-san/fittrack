package desktop

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/vanam-san/fittrack/internal/database"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type DataFilesService struct {
	files     *database.DataFiles
	backupDir string
}

func NewDataFilesService(files *database.DataFiles, backupDir string) *DataFilesService {
	return &DataFilesService{files: files, backupDir: backupDir}
}

func (s *DataFilesService) Export(ctx context.Context, format string) (string, error) {
	if format != "csv" && format != "json" {
		return "", fmt.Errorf("choose CSV or JSON")
	}
	path, err := s.saveDialog("", "fittrack-records."+format, strings.ToUpper(format)+" files", format)
	if err != nil || path == "" {
		return "", err
	}
	return finishExport(path, s.files.Export(ctx, path, format))
}

func (s *DataFilesService) Backup(ctx context.Context) (string, error) {
	dir := s.backupDir
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("open backup folder: %w", err)
	}
	path, err := s.saveDialog(dir, "fittrack-"+time.Now().Format("20060102-150405")+".sqlite3", "FitTrack backups", "sqlite3")
	if err != nil || path == "" {
		return "", err
	}
	return finishExport(path, s.files.Backup(ctx, path))
}

// ChooseInput only selects a path. The UI asks for confirmation before calling Replace.
func (s *DataFilesService) ChooseInput(operation string) (string, error) {
	app := application.Get()
	dialog := app.Dialog.OpenFile().AttachToWindow(app.Window.Current()).CanChooseFiles(true)
	switch operation {
	case "import":
		dialog.SetTitle("Choose a FitTrack export").AddFilter("FitTrack exports", "*.csv;*.json")
	case "restore":
		dialog.SetTitle("Choose a FitTrack backup").AddFilter("FitTrack backups", "*.sqlite3")
		if isDirectory(s.backupDir) {
			dialog.SetDirectory(s.backupDir)
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
	if runtime.GOOS == "android" {
		// The Android document picker returns a content URI. Stage the snapshot
		// privately so the Android host can copy it to that URI.
		root := filepath.Join(filepath.Dir(s.backupDir), "exports")
		if err := os.MkdirAll(root, 0700); err != nil {
			return "", err
		}
		stage, err := os.MkdirTemp(root, "export-")
		if err != nil {
			return "", err
		}
		return filepath.Join(stage, filename), nil
	}
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

func finishExport(path string, err error) (string, error) {
	if err != nil && runtime.GOOS == "android" {
		_ = os.Remove(filepath.Dir(path))
	}
	return path, err
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
