# Run on a disposable Windows user profile. Setup registers FitTrack for that user.
param([string]$PreviousInstaller)

$ErrorActionPreference = 'Stop'
$installer = (Get-ChildItem -Path 'bin/*-installer.exe' | Select-Object -First 1).FullName
if (!$installer) { throw 'Build the per-user NSIS installer first.' }
if ($PreviousInstaller) { $PreviousInstaller = (Resolve-Path -LiteralPath $PreviousInstaller).Path }
if (Test-Path -LiteralPath 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\FitTrackFitTrack') {
    throw 'Use a Windows profile without a registered FitTrack installation.'
}

$root = Join-Path ([IO.Path]::GetTempPath()) ('FitTrack-package-' + [guid]::NewGuid())
$installDir = Join-Path $root 'program'
$executable = Join-Path $installDir 'fittrack.exe'
$uninstaller = Join-Path $installDir 'uninstall.exe'
$originalAppData = $env:LOCALAPPDATA
$env:LOCALAPPDATA = Join-Path $root 'profile'
$database = Join-Path $env:LOCALAPPDATA 'FitTrack/fittrack.sqlite3'
New-Item -ItemType Directory -Path $env:LOCALAPPDATA -Force | Out-Null

function Install-FitTrack {
    param([string]$Package)
    $setup = Start-Process -WindowStyle Hidden -FilePath $Package -ArgumentList "/S /D=$installDir" -PassThru -Wait
    if ($setup.ExitCode -ne 0 -or !(Test-Path -LiteralPath $executable)) {
        throw "FitTrack installer failed with status $($setup.ExitCode)."
    }
}

function Uninstall-FitTrack {
    $setup = Start-Process -WindowStyle Hidden -FilePath $uninstaller -ArgumentList "/S _?=$installDir" -PassThru -Wait
    if ($setup.ExitCode -ne 0 -or (Test-Path -LiteralPath $executable)) {
        throw "FitTrack uninstaller failed with status $($setup.ExitCode)."
    }
    if (Test-Path -LiteralPath $uninstaller) { Remove-Item -LiteralPath $uninstaller }
}

function Open-FitTrack {
    $app = Start-Process -FilePath $executable -PassThru
    try {
        $deadline = [DateTime]::UtcNow.AddSeconds(25)
        do {
            $app.Refresh()
            if ($app.HasExited) { throw "FitTrack quit with status $($app.ExitCode)." }
            if ($app.MainWindowTitle -eq 'FitTrack' -and (Test-Path -LiteralPath $database)) { return }
            Start-Sleep -Milliseconds 250
        } while ([DateTime]::UtcNow -lt $deadline)
        throw 'FitTrack did not open a window and database.'
    } finally {
        if (!$app.HasExited) {
            if (!$app.CloseMainWindow() -or !$app.WaitForExit(5000)) {
                Stop-Process -Id $app.Id -ErrorAction SilentlyContinue
            }
        }
    }
}

try {
    if ($PreviousInstaller) {
        Install-FitTrack $PreviousInstaller
        Open-FitTrack
    } else {
        Install-FitTrack $installer
        Open-FitTrack
    }
    $seed = 'const { DatabaseSync } = require("node:sqlite"); const db = new DatabaseSync(process.argv[1]); db.exec("INSERT INTO app_settings VALUES (1, ''Package check'', ''kg'', ''system'', 0, 0, 0, 0); INSERT INTO weight_measurements (measured_on, weight_kg) VALUES (''2026-11-23'', 80)"); db.close()'
    node -e $seed $database
    if ($LASTEXITCODE -ne 0) { throw 'Could not seed the package test profile.' }
    $before = (Get-FileHash -LiteralPath $database).Hash
    Install-FitTrack $installer
    Open-FitTrack
    if ((Get-FileHash -LiteralPath $database).Hash -ne $before) {
        throw 'Installing over an existing copy modified user data.'
    }

    # An app uninstall must preserve the database and unrelated install-folder files.
    $unrelated = Join-Path $installDir 'keep-me.txt'
    Set-Content -LiteralPath $unrelated -Value 'unrelated file'
    Uninstall-FitTrack
    if (!(Test-Path -LiteralPath $unrelated) -or (Get-FileHash -LiteralPath $database).Hash -ne $before) {
        throw 'Uninstall modified user data.'
    }

    Install-FitTrack $installer
    Open-FitTrack
    if ((Get-FileHash -LiteralPath $database).Hash -ne $before) {
        throw 'Reinstall modified user data.'
    }
    Write-Output 'Windows NSIS install, launch, uninstall and reinstall passed.'
    if ($PreviousInstaller) { Write-Output 'Upgrade from the previous installer preserved FitTrack records.' }
} finally {
    try {
        if (Test-Path -LiteralPath $uninstaller) { Uninstall-FitTrack }
    } finally {
        $env:LOCALAPPDATA = $originalAppData
    }
}
