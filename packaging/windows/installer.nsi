; NSIS Installer Script for DivoomPCMonitorTool
; Requires NSIS 3.0 or later

!define APPNAME "DivoomPCMonitorTool"
!define COMPANYNAME "DivoomPCMonitorTool Team"
!define DESCRIPTION "PC monitoring tool for Divoom devices"
!define VERSIONMAJOR 1
!define VERSIONMINOR 0
!define VERSIONBUILD 0
!define HELPURL "https://github.com/alessio/DivoomPCMonitorTool-Linux"
!define UPDATEURL "https://github.com/alessio/DivoomPCMonitorTool-Linux/releases"
!define ABOUTURL "https://github.com/alessio/DivoomPCMonitorTool-Linux"

RequestExecutionLevel admin
InstallDir "$PROGRAMFILES64\${APPNAME}"
LicenseData "LICENSE"
Name "${APPNAME}"
Icon "icon.ico"
outFile "DivoomPCMonitorTool-Setup.exe"

!include LogicLib.nsh
!include "MUI2.nsh"

!define MUI_ABORTWARNING
!define MUI_ICON "icon.ico"
!define MUI_UNICON "icon.ico"

; Pages
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

; Languages
!insertmacro MUI_LANGUAGE "English"

; Install section
section "install"
    setOutPath $INSTDIR
    
    ; Install files
    file "divoom-monitor-windows.exe"
    file "divoom-daemon-windows.exe" 
    file "divoom-test-windows.exe"
    file "README.txt"
    
    ; Create shortcuts
    createDirectory "$SMPROGRAMS\${APPNAME}"
    createShortCut "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk" "$INSTDIR\divoom-monitor-windows.exe" "" "$INSTDIR\icon.ico"
    createShortCut "$SMPROGRAMS\${APPNAME}\Test Device.lnk" "$INSTDIR\divoom-test-windows.exe" "" "$INSTDIR\icon.ico"
    createShortCut "$SMPROGRAMS\${APPNAME}\Uninstall.lnk" "$INSTDIR\uninstall.exe" "" ""
    
    ; Desktop shortcut
    createShortCut "$DESKTOP\${APPNAME}.lnk" "$INSTDIR\divoom-monitor-windows.exe" "" "$INSTDIR\icon.ico"
    
    ; Create service
    nsExec::ExecToLog '"$INSTDIR\divoom-daemon-windows.exe" install'
    
    ; Registry entries for Add/Remove Programs
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayName" "${APPNAME} - ${DESCRIPTION}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "UninstallString" "$\"$INSTDIR\uninstall.exe$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "QuietUninstallString" "$\"$INSTDIR\uninstall.exe$\" /S"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "InstallLocation" "$\"$INSTDIR$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayIcon" "$\"$INSTDIR\icon.ico$\""
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "Publisher" "${COMPANYNAME}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "HelpLink" "${HELPURL}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "URLUpdateInfo" "${UPDATEURL}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "URLInfoAbout" "${ABOUTURL}"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "DisplayVersion" "${VERSIONMAJOR}.${VERSIONMINOR}.${VERSIONBUILD}"
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "VersionMajor" ${VERSIONMAJOR}
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "VersionMinor" ${VERSIONMINOR}
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}" "NoRepair" 1
    
    writeUninstaller "$INSTDIR\uninstall.exe"
sectionEnd

; Uninstall section
section "uninstall"
    ; Stop and remove service
    nsExec::ExecToLog '"$INSTDIR\divoom-daemon-windows.exe" stop'
    nsExec::ExecToLog '"$INSTDIR\divoom-daemon-windows.exe" remove'
    
    ; Remove files
    delete "$INSTDIR\divoom-monitor-windows.exe"
    delete "$INSTDIR\divoom-daemon-windows.exe"
    delete "$INSTDIR\divoom-test-windows.exe"
    delete "$INSTDIR\README.txt"
    delete "$INSTDIR\icon.ico"
    delete "$INSTDIR\uninstall.exe"
    
    ; Remove shortcuts
    delete "$SMPROGRAMS\${APPNAME}\${APPNAME}.lnk"
    delete "$SMPROGRAMS\${APPNAME}\Test Device.lnk"
    delete "$SMPROGRAMS\${APPNAME}\Uninstall.lnk"
    rmDir "$SMPROGRAMS\${APPNAME}"
    delete "$DESKTOP\${APPNAME}.lnk"
    
    ; Remove registry entries
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APPNAME}"
    
    ; Remove directory if empty
    rmDir "$INSTDIR"
sectionEnd