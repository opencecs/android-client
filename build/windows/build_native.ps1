param(
    [string]$BuildFlags,
    [string]$BinDir,
    [string]$AppName,
    [string]$CgoEnabled = "1",
    [string]$Arch = "amd64",
    [string]$Dev = "false"
)

# 直接从 build/config.yml 读版本号，不依赖 python
$configContent = Get-Content "build/config.yml" -Raw -EA SilentlyContinue
$versionMatch = [regex]::Match($configContent, 'version:\s*"?([0-9]+\.[0-9]+\.[0-9]+)')
$appVersion = if ($versionMatch.Success) { $versionMatch.Groups[1].Value } else { "dev" }
$buildTime = (Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ")
$ldflags = "-X edgeclient/updater.AppVersion=$appVersion -X edgeclient/updater.BuildTime=$buildTime -w -s -H windowsgui"

if ($Dev -eq "true") {
    $ldflags = $ldflags -replace '\s*-w\b', ''
    $ldflags = $ldflags -replace '\s*-s\b', ''
    $ldflags = $ldflags -replace '\s*-H\s+windowsgui\b', ''
    $ldflags = $ldflags.Trim()
}

$env:GOOS = "windows"
$env:CGO_ENABLED = $CgoEnabled
$env:GOARCH = $Arch
# 优先使用 WinLibs（winget 安装），回退到 Chocolatey
$winlibsBase = "$env:LOCALAPPDATA\Microsoft\WinGet\Packages"
$winlibsBin = (Get-ChildItem $winlibsBase -Filter "g++.exe" -Recurse -EA SilentlyContinue | Select-Object -First 1).DirectoryName
if ($winlibsBin) {
    $env:CC = "$winlibsBin\gcc.exe"
    $env:CXX = "$winlibsBin\g++.exe"
    $env:PATH = "$winlibsBin;" + $env:PATH
}
else {
    $env:CC = "C:\ProgramData\chocolatey\bin\gcc.exe"
    $env:CXX = "C:\ProgramData\chocolatey\bin\g++.exe"
    $env:PATH = "C:\ProgramData\chocolatey\bin;" + $env:PATH
}
# 静态链接 MinGW 运行时，实现单文件运行，无需附带 libgcc/libstdc++/libwinpthread DLL
$env:CGO_LDFLAGS = "-static-libgcc -static-libstdc++ -Wl,-Bstatic -lwinpthread -Wl,-Bdynamic"

$outPath = "$BinDir/$AppName.exe"

Write-Host "Dev    : $Dev"
Write-Host "LDFLAGS: $ldflags"
Write-Host "Output : $outPath"

# 根据 Dev 模式直接构建 go build 参数，避免 PowerShell 引号传递问题
if ($Dev -eq "true") {
    $goArgs = @("-p", "20", "-tags", "cgorpa", "-buildvcs=false", "-gcflags=all=-l", "-ldflags=$ldflags", "-o", $outPath)
}
else {
    $goArgs = @("-p", "20", "-tags", "cgorpa production devtools", "-trimpath", "-buildvcs=false", "-ldflags=$ldflags", "-o", $outPath)
}

& go build @goArgs
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "Build complete (static MinGW runtime, no external DLLs needed)"
exit 0
