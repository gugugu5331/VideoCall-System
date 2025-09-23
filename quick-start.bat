@echo off
chcp 65001 >nul
title VideoCall System - Quick Start

echo ==========================================
echo 🚀 VideoCall System - Quick Start
echo ==========================================
echo.

echo Starting system...
cd /d "%~dp0"
call scripts\startup\start_system_simple.bat

echo.
echo ✅ Quick start completed!
pause
