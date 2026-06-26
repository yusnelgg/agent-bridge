@echo off
chcp 65001 >nul
set TARGET_DIR=%USERPROFILE%\agent-bridge
mkdir "%TARGET_DIR%\bin" "%TARGET_DIR%\configs" "%TARGET_DIR%\prompts" 2>nul
copy /Y "%~dp0bin\agent.exe" "%TARGET_DIR%\bin\agent.exe" >nul
copy /Y "%~dp0bin\agent-bridge.exe" "%TARGET_DIR%\bin\agent-bridge.exe" >nul
copy /Y "%~dp0configs\frontend.yaml" "%TARGET_DIR%\configs\frontend.yaml" >nul
copy /Y "%~dp0configs\backend.yaml" "%TARGET_DIR%\configs\backend.yaml" >nul
copy /Y "%~dp0AGENTS.md" "%TARGET_DIR%\AGENTS.md" >nul
copy /Y "%~dp0prompts\frontend-ai.md" "%TARGET_DIR%\prompts\frontend-ai.md" >nul
copy /Y "%~dp0prompts\backend-ai.md" "%TARGET_DIR%\prompts\backend-ai.md" >nul
for /f "tokens=2*" %%A in ('reg query HKCU\Environment /v Path 2^>nul ^| find "Path"') do set OLD_PATH=%%B
echo !OLD_PATH! | find /i "%TARGET_DIR%\bin" >nul
if %errorlevel% neq 0 setx PATH "%TARGET_DIR%\bin;!OLD_PATH!"
echo Installed. Restart your terminal.
pause
