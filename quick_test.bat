@echo off
chcp 65001 >nul
title VideoCall System - Quick Test

echo ==========================================
echo VideoCall System - Quick Test
echo ==========================================
echo.

echo Running system tests...
cd /d "%~dp0"
python scripts\testing\run_all_tests.py

echo.
echo Test completed!
pause
