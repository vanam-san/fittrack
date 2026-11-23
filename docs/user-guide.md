# FitTrack user guide

## Install and run offline

| Package | Prerequisites and installation |
| --- | --- |
| Windows x64 installer | Use 64-bit Windows 10 or 11 with the Microsoft Edge WebView2 Evergreen Runtime. If needed, install Microsoft's x64 Evergreen Standalone Installer first. Run FitTrack Setup and open FitTrack from Start. [Microsoft WebView2 downloads](https://developer.microsoft.com/microsoft-edge/webview2/#download-section) |
| Windows x64 executable | With the same prerequisites, keep the executable in a writable folder and open it directly. It shares the usual FitTrack data directory for the Windows account. |
| macOS universal disk image | Use macOS 12 or newer on Apple Silicon or Intel. Open the disk image, drag FitTrack to Applications, and open it there. For distribution outside development, use a signed and notarized disk image. |
| Linux amd64 Debian package | Use Ubuntu 24.04 or Debian 13 with GTK 4 and WebKitGTK 6.0 installed. Install with `sudo dpkg -i ./fittrack.deb` and launch FitTrack from the applications menu. Place the package and any missing dependency packages on the machine before disconnecting. [Wails platform prerequisites](https://v3.wails.io/quick-start/installation/) |

Setup, logging, analytics, export, import, and recovery work from the installed app with the network disconnected. On first launch, set your preferences or open **Data and recovery** on the setup screen to import records or restore a backup.

The Windows installer includes a **FitTrack storage** page. Choose the database and backup folders independently, or leave either field blank to use its default. To transfer or recover your data, select a FitTrack `.sqlite3` backup and confirm the restore. FitTrack validates and restores the file on its first launch. If the restore fails, open **Data and recovery** to retry.

Storage choices apply to the Windows account running Setup. Close FitTrack before rerunning Setup to change folders. If you choose a different database folder, the original database stays in its current folder. Select a backup during Setup to transfer its records and preferences to the new folder.

## Record and review your activity

- **Workouts** lets you build routines, start a workout, log sets, and resume an unfinished session after closing FitTrack.
  - Open **History → Browse calendar** to find highlighted workout dates. Select a day to see its completed sessions, or choose **Show all dates** to clear the filter. Within a saved workout, use **Newer workout** and **Older workout** to move through the current results.
  - In **Exercise library → Custom exercise**, choose **Save and add another** to save the exercise and open a blank form.
- **Nutrition** lets you log meals, snacks, and drinks, use food presets, and edit entries by date.
- **Weight** lets you record measurements and a target and review a date range.
- **Analytics** groups workout, nutrition, and weight results for the date range you select.
- Use **Tab** and **Shift+Tab** to move through controls. Press **Enter** or **Space** to activate focused buttons. Use **Esc** to dismiss a file picker or confirmation before committing a change. After selecting a navigation item, FitTrack focuses the page content.

## Find your files

Open **Settings → Data and recovery** for the exact paths. The active database is `fittrack.sqlite3` and the default backup folder is `backups` within the same data directory.

| System | Data directory |
| --- | --- |
| Windows | `%LOCALAPPDATA%\FitTrack` |
| macOS | `~/Library/Application Support/FitTrack` |
| Linux | `$XDG_DATA_HOME/FitTrack` when set to an absolute path, otherwise `~/.local/share/FitTrack` |

**Create backup** starts in the configured backup folder and lets you choose a different destination. FitTrack shows the saved path when it finishes. Copy an important backup to a separate drive to recover after a device or disk failure.

## Transfer records between installations

1. On the old installation, choose **Export CSV** or **Export JSON** and save the file.
2. On the destination, choose **Import records**, select the exported file, and confirm.
3. Verify your preferences and records. Import replaces the destination's settings and records, including exercise IDs, dated measurements, food entries, and the current workout. FitTrack saves a `pre-import-*.sqlite3` recovery backup first.

Both versions of the export include the same records. JSON is UTF-8 with a top-level `version` of `1` and a `tables` object containing named arrays of records. Each record uses the exported database field names. CSV is UTF-8 with one fixed header. Each row identifies its table in `table` and has `fittrack_version` set to `1`. Unused columns are blank. A final `__end__` row holds the number of data rows in `record_count`. Ordinary data rows leave that column blank. CSV text escapes literal backslashes as `\\` and carriage returns as `\r` so line endings round-trip. Both exports use calendar dates in `YYYY-MM-DD`, kilograms for weights, metres for distances, and seconds for set duration. Linked records use the exported IDs. A food log without a preset has an empty CSV `food_id` or a JSON `null`.

Keep the original export for transfer. Spreadsheet apps can change text, dates, and long IDs when saving. An invalid file or broken reference stops import and the existing records remain available.

## Back up and recover

1. Choose **Create backup** for a full SQLite snapshot, including schema metadata.
2. To recover, choose **Restore backup**, select the `.sqlite3` file, and confirm the filename shown.
3. FitTrack validates the backup, prepares supported older backups for this version, and saves a `pre-restore-*.sqlite3` copy of the current database in the configured backup folder. Restore replaces current records and preferences. If a restore fails during replacement, the existing database and its recovery copy remain available.

Close FitTrack before upgrading or uninstalling. On Windows, run the newer installer with the same install scope or replace the portable executable. On macOS, replace the Applications bundle. On Linux, install the newer Debian package over the existing one. Standard uninstall removes the app. A reinstall can reuse your database and backups in their existing folders.

If FitTrack cannot open the current database, quit all FitTrack windows. Move the data directory to a separate recovery location so the original database, sidecar files, and backups stay together. Relaunch FitTrack to create a fresh database, then restore a known-good backup from the moved folder or a separate drive. If FitTrack reports a newer backup schema, use the same or a newer FitTrack release than the one that created it.
