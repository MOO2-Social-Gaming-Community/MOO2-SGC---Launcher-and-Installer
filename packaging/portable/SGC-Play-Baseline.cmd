@echo off
setlocal
set "ROOT=%~dp0."
set "SETUP=%ROOT%\MOO2-SGC-Setup.exe"
if not exist "%SETUP%" (
  echo MOO2-SGC-Setup.exe is missing. Extract the complete integration kit here.
  pause
  exit /b 1
)
"%SETUP%" --install-root "%ROOT%" --portable --command ensure-installed --offline "%ROOT%\distribution\offline-0.4.5" --no-launch
if errorlevel 1 goto failed
"%SETUP%" --install-root "%ROOT%" --portable --command launch-installed --launcher-command play --profile baseline
if errorlevel 1 goto failed
exit /b 0
:failed
echo.
echo The operation stopped. Keep the exact error above. Do not disable protection.
echo Close the launcher and game before running another management command.
echo Files and retained backups are not deleted by this script.
pause
exit /b 1
