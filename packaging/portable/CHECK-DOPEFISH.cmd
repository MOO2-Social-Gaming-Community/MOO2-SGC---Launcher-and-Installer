@echo off
setlocal DisableDelayedExpansion
set "ROOT=%~dp0."
set "SETUP=%ROOT%\MOO2-SGC-Setup.exe"
if not exist "%SETUP%" (
  echo Extract the complete 0.4.7 portable integration kit here first.
  pause
  exit /b 1
)
echo MOO2-SGC 0.4.7 - Dopefish connection check - no game
 echo Close the launcher and MOO2 first, or use the connection-check button in the launcher.
echo This contacts moo2.thedopefish.com over UDP 213 through your verified DOSBox.
echo No game is mounted or launched. Read CONNECT / STATUS. Type EXIT to finish.
choice /C YN /N /M "Start this network diagnostic now? [Y/N] "
if errorlevel 2 exit /b 0
"%SETUP%" --install-root "%ROOT%" --portable --command ensure-installed --offline "%ROOT%\distribution\offline-0.4.7" --no-launch
if errorlevel 1 goto failed
"%SETUP%" --install-root "%ROOT%" --portable --command launch-installed --launcher-command check-dopefish --no-browser
if errorlevel 1 goto failed
pause
exit /b 0
:failed
echo Diagnostic stopped. Keep the error above; do not remove an active installation lock.
pause
exit /b 1
