# Native release checks

FitTrack 0.2.0 adds CSV and JSON transfer, SQLite snapshots, confirmed recovery, and a shared FitTrack icon in the app and installers. Package metadata supports upgrading from 0.1.0. The Windows installer uses a preinstalled WebView2 runtime and preserves unrelated files in the installation folder on uninstall.

Build with the dependencies in the README and the platform tools below. Run each build on its target operating system.

| Target | Build command | Artifact |
| --- | --- | --- |
| Windows x64 with NSIS on PATH | `wails3 package INSTALL_SCOPE=user` | `bin/FitTrack-amd64-installer.exe` and the standalone `bin/fittrack.exe` |
| macOS with Xcode command-line tools | `wails3 task darwin:package:universal`, then `wails3 task darwin:create:dmg` | Disk image under `bin/` |
| Linux amd64 with GTK 4 and WebKitGTK 6.0 development headers | `wails3 task linux:create:deb` | Debian package under `bin/` |

The [native workflow](../.github/workflows/native-release.yml) builds these artifacts on Windows, macOS, and Linux. It runs Go and frontend checks on each runner, tests Windows installer lifecycle and WebView workflows, verifies and launches the universal macOS bundle, and runs Debian installer lifecycle with app launch in a network namespace. The workflow artifacts are for qualification. Apply appropriate production signing and notarization before distribution.

## Release acceptance

Run the following on a clean native desktop account for each artifact, including the standalone Windows executable. Keep a 0.1.0 package and the platform runtime dependencies available locally. Disconnect before installing FitTrack and keep the machine offline for the run.

| Scenario | Pass condition |
| --- | --- |
| Fresh install | Launch from the installer shortcut, Applications, apps menu, or standalone file. Complete setup. Create a workout routine, pause and resume a logged set, complete the workout, add nutrition and weight entries, set a goal, and view analytics. Quit and reopen to verify persistence. |
| Upgrade | On a separate profile, populate version 0.1.0, quit it, and install or replace the app with 0.2.0. Confirm the earlier profile, workout history, nutrition, weight, and goals still render. |
| Uninstall and reinstall | Remove the app with the platform's usual uninstall process, then reinstall. Verify the database and backups are still present and the app loads the same records. For the standalone executable, remove and replace the executable. |
| Portable records | Export both formats to a custom directory. Import each into its own clean profile and verify preferences and all tracking data. Attempt an import with an invalid record and recheck the original records. |
| Recovery | Create a SQLite backup, change records, and restore the backup. Cancel once at the confirmation and verify the change remains. Confirm on the second attempt and verify the earlier state and a newly created recovery backup. Try a corrupt backup and verify the current records remain available. |
| Keyboard and dialogs | Use Tab, Shift+Tab, Enter, Space, and Esc in setup, Settings, navigation, file pickers, and destructive confirmations. Verify focus remains visible, Cancel keeps records, and all file actions can be completed by keyboard. |
| Offline | Verify local logs, charts, import, export, and recovery work for the entire run while disconnected. |

Record the operating system version, architecture, artifact SHA-256, and results for each scenario before releasing. Repeat the macOS run on Apple Silicon and Intel for the universal image. For Windows upgrades use the same installer scope across versions.
