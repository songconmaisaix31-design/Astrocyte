@echo off
setlocal
powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0Start-Astrocyte.ps1" %*
set "release_exit=%errorlevel%"
if not "%release_exit%"=="0" pause
exit /b %release_exit%
