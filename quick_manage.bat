@echo off
chcp 65001 >nul
title VideoCall System - Quick Manage

echo ==========================================
echo VideoCall System - Quick Manage
echo ==========================================
echo.

echo Opening management menu...
cd /d "%~dp0"
call scripts\management\manage_system.bat

echo.
echo Management completed!
pause
