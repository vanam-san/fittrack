# Native release checks

The Windows installer uses a preinstalled WebView2 runtime and preserves unrelated files in the installation folder on uninstall.

Build with the dependencies in the README and the platform tools below. Run each build on its target operating system.

| Target | Build command | Artifact |
| --- | --- | --- |
| Windows x64 with NSIS on PATH | `wails3 package INSTALL_SCOPE=user` | `bin/FitTrack-amd64-installer.exe` and the standalone `bin/fittrack.exe` |
| macOS with Xcode command-line tools | `wails3 task darwin:package:universal`, then `wails3 task darwin:create:dmg` | Disk image under `bin/` |
| Linux amd64 with GTK 4 and WebKitGTK 6.0 development headers | `wails3 task linux:create:deb` | Debian package under `bin/` |

The [native workflow](../.github/workflows/native-release.yml) builds these artifacts on Windows, macOS, and Linux. It runs Go and frontend checks on each runner, tests Windows installer lifecycle and WebView workflows, verifies and launches the universal macOS bundle, and runs Debian installer lifecycle with app launch in a network namespace. The workflow does not perform production signing or macOS notarization.

## Publish a release

Start with `v1.0.0` and use stable `vMAJOR.MINOR.PATCH` tags.

| Change | Version increment | Example |
| --- | --- | --- |
| Major overhaul or breaking change | Increase major and reset minor and patch to zero | `v1.4.2` → `v2.0.0` |
| Backward-compatible feature | Increase minor and reset patch to zero | `v1.0.0` → `v1.1.0` |
| Backward-compatible fix | Increase patch | `v1.1.0` → `v1.1.1` |

Before tagging, commit the release notes to `docs/releases/<tag>.md` with
`## Added` and `## Changed` sections. After completing release acceptance, tag
the release commit and push the tag:

```sh
git tag -a v1.0.0 -m "FitTrack v1.0.0"
git push origin refs/tags/v1.0.0
```

GitHub Actions checks the tag format and uses its version in the package metadata.
When all three platform jobs pass, it publishes the GitHub Release with the
notes from the tagged commit, the Windows x64 installer and standalone executable,
the macOS universal disk image, and the Debian amd64 installer. Branch builds and pull requests
continue to upload workflow artifacts.

Use the **Native release** run for the tag to check each platform job and the
final publish job. Once it succeeds, open the release page and check that all four
downloads are listed. Keep the published tag attached to its original commit.
If a published build needs a fix, use the next patch tag.

## Diagnose package checks

- **Debian** runs the installed executable as the runner user under Xvfb and
  D-Bus. A private network namespace keeps the test offline while allowing
  loopback. The test expects FitTrack to initialize a database under a temporary
  `XDG_DATA_HOME` and stay running until the timeout. It then writes sample data
  and checks that uninstalling and reinstalling preserves the database. GitHub's
  Ubuntu runner blocks WebKit’s nested namespace, so the test launch disables
  WebKit’s sandbox inside the offline namespace. The distributed app uses the
  platform’s normal WebKit sandbox settings.
- **macOS** verifies the disk image and the copied app’s signature, checks the
  app binary for both `x86_64` and `arm64`, and launches the app twice to check
  database preservation.
- **Windows** uses the per-user NSIS installer and runs a WebView smoke test with
  a temporary profile. The package check covers installation, uninstallation,
  and reinstall.

## Release acceptance

Run the following on a clean native desktop account for each artifact, including the standalone Windows executable. Keep a previous package and the platform runtime dependencies available locally. Disconnect before installing FitTrack and keep the machine offline for the run.

| Scenario | Pass condition |
| --- | --- |
| Fresh install | Launch from the installer shortcut, Applications, apps menu, or standalone file. Complete setup. Create a workout routine, pause and resume a logged set, complete the workout, add nutrition and weight entries, set a goal, and view analytics. Quit and reopen to verify persistence. |
| Upgrade | On a separate profile, populate the previous version, quit it, and install or replace the app with the candidate version. Confirm the earlier profile, workout history, nutrition, weight, and goals still render. |
| Uninstall and reinstall | Remove the app with the platform's usual uninstall process, then reinstall. Verify the database and backups are still present and the app loads the same records. For the standalone executable, remove and replace the executable. |
| Portable records | Export both formats to a custom directory. Import each into its own clean profile and verify preferences and all tracking data. Attempt an import with an invalid record and recheck the original records. |
| Recovery | Create a SQLite backup, change records, and restore the backup. Cancel once at the confirmation and verify the change remains. Confirm on the second attempt and verify the earlier state and a newly created recovery backup. Try a corrupt backup and verify the current records remain available. |
| Keyboard and dialogs | Use Tab, Shift+Tab, Enter, Space, and Esc in setup, Settings, navigation, file pickers, and destructive confirmations. Verify focus remains visible, Cancel keeps records, and all file actions can be completed by keyboard. |
| Offline | Verify local logs, charts, import, export, and recovery work for the entire run while disconnected. |

Record the operating system version, architecture, artifact SHA-256, and results for each scenario before releasing. Repeat the macOS run on Apple Silicon and Intel for the universal image. For Windows upgrades use the same installer scope across versions.
