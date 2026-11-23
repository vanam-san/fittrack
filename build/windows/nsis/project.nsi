Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows you to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
## 
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the wails_tools.nsh file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "my-project" # Default "FitTrack"
## !define INFO_COMPANYNAME    "FitTrack" # Default "FitTrack"
## !define INFO_PRODUCTNAME    "My Product Name" # Default "FitTrack"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "0.2.0"
## !define INFO_COPYRIGHT      "(c) Now, FitTrack" # Default "© 2026, FitTrack"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
## !define WAILS_INSTALL_SCOPE     "user"             # Default "machine" - set to "user" for per-user install ($LOCALAPPDATA) without UAC prompt
####
## Include the wails tools
####
!include "wails_tools.nsh"

# The evergreen standalone runtime can be installed before FitTrack while offline.
!macro fittrack.requireWebView2
    SetRegView 64
    ReadRegStr $0 HKLM "SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
    ${If} $0 == ""
    ${OrIf} $0 == "0.0.0.0"
        ReadRegStr $0 HKCU "Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" "pv"
    ${EndIf}
    ${If} $0 == ""
    ${OrIf} $0 == "0.0.0.0"
        MessageBox MB_OK|MB_ICONSTOP "Install the Microsoft Edge WebView2 Evergreen Runtime before installing FitTrack." /SD IDOK
        SetErrorLevel 2
        Quit
    ${EndIf}
!macroend

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"
!include "nsDialogs.nsh"

!define FITTRACK_STORAGE_KEY "Software\FitTrack\Storage"
Var DatabaseDirectory
Var BackupDirectory
Var RestoreSource
Var DatabaseField
Var BackupField
Var RestoreField

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
Page custom StoragePage StoragePageLeave
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uninstalling page

!insertmacro MUI_LANGUAGE "English" # Set the Language of the installer

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
!if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
!else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif
ShowInstDetails show # This will always show the installation details.

Function .onInit
   !insertmacro wails.checkArchitecture
   SetRegView 64
   ReadRegStr $DatabaseDirectory HKCU "${FITTRACK_STORAGE_KEY}" "DatabaseDirectory"
   ReadRegStr $BackupDirectory HKCU "${FITTRACK_STORAGE_KEY}" "BackupDirectory"
FunctionEnd

!macro FitTrackStorageField label top field value callback
    ${NSD_CreateLabel} 0 ${top}u 100% 10u "${label}"
    Pop $0
    !define /math fieldTop ${top} + 11
    ${NSD_CreateText} 0 ${fieldTop}u 77% 13u "${value}"
    Pop ${field}
    ${NSD_CreateBrowseButton} 79% ${fieldTop}u 21% 13u "Browse..."
    Pop $0
    ${NSD_OnClick} $0 ${callback}
    !undef fieldTop
!macroend

Function StoragePage
    !insertmacro MUI_HEADER_TEXT "FitTrack storage" "Choose where FitTrack keeps your data."
    nsDialogs::Create 1018
    Pop $0
    ${If} $0 == error
        Abort
    ${EndIf}

    ${NSD_CreateLabel} 0 0 100% 25u "Leave folders blank to use the defaults. Choosing a different database folder leaves any existing database in its current folder."
    Pop $0
    !insertmacro FitTrackStorageField "Database folder (default: Local AppData\FitTrack)" 29 $DatabaseField $DatabaseDirectory BrowseDatabase
    !insertmacro FitTrackStorageField "Backup folder (default: backups inside the database folder)" 59 $BackupField $BackupDirectory BrowseBackups
    !insertmacro FitTrackStorageField "Restore from a FitTrack backup on first launch (optional)" 89 $RestoreField $RestoreSource BrowseRestore
    nsDialogs::Show
FunctionEnd

Function BrowseDatabase
    ${NSD_GetText} $DatabaseField $0
    nsDialogs::SelectFolderDialog "Choose the database folder" "$0"
    Pop $0
    ${If} $0 != error
        ${NSD_SetText} $DatabaseField "$0"
    ${EndIf}
FunctionEnd

Function BrowseBackups
    ${NSD_GetText} $BackupField $0
    nsDialogs::SelectFolderDialog "Choose the backup folder" "$0"
    Pop $0
    ${If} $0 != error
        ${NSD_SetText} $BackupField "$0"
    ${EndIf}
FunctionEnd

Function BrowseRestore
    nsDialogs::SelectFileDialog open "" "FitTrack backups|*.sqlite3"
    Pop $0
    ${If} $0 != error
        ${NSD_SetText} $RestoreField "$0"
    ${EndIf}
FunctionEnd

!macro FitTrackCheckPath value label
    ${If} ${value} != ""
        System::Call 'shlwapi::PathIsRelativeW(w "${value}") i.r0'
        StrCpy $1 "${value}" 2
        StrCpy $2 "${value}" 1 2
        ${If} $0 != 0
            MessageBox MB_OK|MB_ICONEXCLAMATION "Enter a full path for the ${label} or use Browse."
            Abort
        ${EndIf}
        ${If} $1 != "\\"
            StrCpy $1 "${value}" 1 1
            ${If} $1 != ":"
                StrCpy $2 ""
            ${EndIf}
            ${If} $2 != "\"
            ${AndIf} $2 != "/"
                MessageBox MB_OK|MB_ICONEXCLAMATION "Enter a full path for the ${label} or use Browse."
                Abort
            ${EndIf}
        ${EndIf}
    ${EndIf}
!macroend

Function StoragePageLeave
    ${NSD_GetText} $DatabaseField $DatabaseDirectory
    ${NSD_GetText} $BackupField $BackupDirectory
    ${NSD_GetText} $RestoreField $RestoreSource
    !insertmacro FitTrackCheckPath $DatabaseDirectory "database folder"
    !insertmacro FitTrackCheckPath $BackupDirectory "backup folder"
    !insertmacro FitTrackCheckPath $RestoreSource "restore file"

    ${If} $RestoreSource != ""
        ${IfNot} ${FileExists} "$RestoreSource"
            MessageBox MB_OK|MB_ICONEXCLAMATION "Choose an existing FitTrack backup file."
            Abort
        ${EndIf}
        MessageBox MB_OKCANCEL|MB_ICONEXCLAMATION "On first launch, FitTrack will replace any records and preferences in the selected database with this backup." IDOK +2
        Abort
    ${EndIf}
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro fittrack.requireWebView2

    SetOutPath $INSTDIR
    
    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols
    
    !insertmacro wails.writeUninstaller

    ${IfNot} ${Silent}
        SetRegView 64
        ClearErrors
        WriteRegStr HKCU "${FITTRACK_STORAGE_KEY}" "DatabaseDirectory" "$DatabaseDirectory"
        WriteRegStr HKCU "${FITTRACK_STORAGE_KEY}" "BackupDirectory" "$BackupDirectory"
        WriteRegStr HKCU "${FITTRACK_STORAGE_KEY}" "RestoreSource" "$RestoreSource"
        ${If} ${Errors}
            MessageBox MB_OK|MB_ICONSTOP "FitTrack could not save the selected storage settings."
            SetErrorLevel 3
            Abort
        ${EndIf}
    ${EndIf}
SectionEnd

Section "uninstall" 
    !insertmacro wails.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    Delete "$INSTDIR\${PRODUCT_EXECUTABLE}"

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
    RMDir "$INSTDIR"
SectionEnd
