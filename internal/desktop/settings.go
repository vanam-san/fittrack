package desktop

import (
	"context"
	"errors"
	"log/slog"

	"github.com/vanam-san/fittrack/internal/config"
	"github.com/vanam-san/fittrack/internal/settings"
)

type SettingsState struct {
	Preferences   settings.Preferences `json:"preferences"`
	SetupComplete bool                 `json:"setupComplete"`
	DatabasePath  string               `json:"databasePath"`
	BackupPath    string               `json:"backupPath"`
	RestoreError  string               `json:"restoreError,omitempty"`
}

type SettingsService struct {
	settings     *settings.Service
	storage      config.Storage
	restoreError string
}

func NewSettingsService(service *settings.Service, storage config.Storage, restoreError string) *SettingsService {
	return &SettingsService{settings: service, storage: storage, restoreError: restoreError}
}

func (s *SettingsService) Load(ctx context.Context) (SettingsState, error) {
	if s.settings == nil {
		return SettingsState{}, errors.New("FitTrack couldn't open its local database. Check that the selected database folder is writable, then restart FitTrack.")
	}
	preferences, complete, err := s.settings.Load(ctx)
	if err != nil {
		slog.Error("could not load local settings")
		return SettingsState{}, errors.New("FitTrack couldn't load your settings. Restart FitTrack and try again.")
	}
	return s.state(preferences, complete), nil
}

func (s *SettingsService) Save(ctx context.Context, preferences settings.Preferences) (SettingsState, error) {
	if s.settings == nil {
		return SettingsState{}, errors.New("Local storage is unavailable. Restart FitTrack and try again.")
	}
	saved, err := s.settings.Save(ctx, preferences)
	if err != nil {
		var validation settings.ValidationError
		if errors.As(err, &validation) {
			return SettingsState{}, validation
		}
		slog.Error("could not save local settings")
		return SettingsState{}, errors.New("FitTrack couldn't save your settings. Try again.")
	}
	return s.state(saved, true), nil
}

func (s *SettingsService) state(preferences settings.Preferences, complete bool) SettingsState {
	return SettingsState{
		Preferences: preferences, SetupComplete: complete,
		DatabasePath: s.storage.DatabasePath, BackupPath: s.storage.BackupPath,
		RestoreError: s.restoreError,
	}
}
