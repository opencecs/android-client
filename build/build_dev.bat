@echo off
set PATH=%LOCALAPPDATA%\Microsoft\WinGet\Packages\BrechtSanders.WinLibs.POSIX.UCRT_Microsoft.Winget.Source_8wekyb3d8bbwe\mingw64\bin;%PATH%
set CGO_ENABLED=1
wails3 build DEV=true BUILD_FLAGS="-tags cgorpa" CGO_ENABLED=1
