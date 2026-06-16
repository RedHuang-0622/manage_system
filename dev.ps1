# Lab Management System — Development Launcher
# Usage: .\dev.ps1          (both backend + frontend)
#        .\dev.ps1 backend   (backend only)
#        .\dev.ps1 frontend  (frontend only)

param([string]$target = "both")

$ErrorActionPreference = "Stop"
$backendDir = "$PSScriptRoot\backend"
$frontendDir = "$PSScriptRoot\frontend"
$backendPort = 8080
$frontendPort = 5173

# ── Helper: Kill processes on a port ─────────────────────────────
function Stop-Port($port) {
    $conns = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue
    if ($conns) {
        Write-Host "[stop] Killing processes on port $port..." -ForegroundColor Yellow
        $conns | ForEach-Object {
            Stop-Process -Id $_.OwningProcess -Force -ErrorAction SilentlyContinue
        }
    }
}

# ── Helper: Wait for a port to be listening ──────────────────────
function Wait-Port($port, $timeoutSec = 15) {
    $elapsed = 0
    while ($elapsed -lt $timeoutSec) {
        $listening = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue
        if ($listening) { return $true }
        Start-Sleep -Seconds 1
        $elapsed++
    }
    return $false
}

# ── Resolve executable path ─────────────────────────────────────
function Find-Exe($name) {
    $cmd = Get-Command $name -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
    # Fallback: common install paths
    $paths = @(
        "$env:LOCALAPPDATA\Programs\Go\bin\$name.exe",
        "C:\Program Files\Go\bin\$name.exe",
        "C:\go\bin\$name.exe",
        "$env:APPDATA\npm\$name.cmd"
    )
    foreach ($p in $paths) {
        if (Test-Path $p) { return $p }
    }
    throw "Cannot find $name in PATH or common locations"
}

# ── Start Backend ────────────────────────────────────────────────
function Start-Backend {
    Write-Host ""
    Write-Host "=== [1/2] Starting backend ===" -ForegroundColor Cyan
    Write-Host "  URL: http://localhost:$backendPort" -ForegroundColor Gray

    Stop-Port $backendPort

    $goExe = Find-Exe "go"
    Write-Host "  go: $goExe" -ForegroundColor DarkGray

    # Use .NET ProcessStartInfo for reliable background launch
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = $goExe
    $psi.Arguments = "run ./cmd/main.go"
    $psi.WorkingDirectory = $backendDir
    $psi.UseShellExecute = $false
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.Environment["PATH"] = $env:PATH

    $proc = [System.Diagnostics.Process]::Start($psi)

    Write-Host "  Waiting for backend to be ready..." -ForegroundColor Gray
    $ready = Wait-Port $backendPort 15
    if (-not $ready) {
        Write-Host "  ERROR: Backend failed to start within 15s" -ForegroundColor Red
        $proc.Kill()
        exit 1
    }
    Write-Host "  Backend ready (PID: $($proc.Id))" -ForegroundColor Green
    return $proc
}

# ── Start Frontend ───────────────────────────────────────────────
function Start-Frontend {
    Write-Host ""
    Write-Host "=== [2/2] Starting frontend ===" -ForegroundColor Cyan
    Write-Host "  URL: http://localhost:$frontendPort" -ForegroundColor Gray

    Stop-Port $frontendPort

    # Check node_modules
    if (-not (Test-Path "$frontendDir\node_modules")) {
        Write-Host "  Installing dependencies..." -ForegroundColor Yellow
        & npm --prefix $frontendDir install
    }

    Write-Host "  npx: npx.cmd" -ForegroundColor DarkGray

    # Use ShellExecute for .cmd files (must run through cmd.exe)
    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = "cmd.exe"
    $psi.Arguments = "/c `"cd /d $frontendDir && npx vite --host`""
    $psi.WorkingDirectory = $frontendDir
    $psi.UseShellExecute = $false
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true

    $proc = [System.Diagnostics.Process]::Start($psi)

    Write-Host "  Waiting for frontend to be ready..." -ForegroundColor Gray
    $ready = Wait-Port $frontendPort 10
    if (-not $ready) {
        Write-Host "  ERROR: Frontend failed to start within 10s" -ForegroundColor Red
        $proc.Kill()
        exit 1
    }
    Write-Host "  Frontend ready (PID: $($proc.Id))" -ForegroundColor Green
    return $proc
}

# ── Main ─────────────────────────────────────────────────────────
$backendProc = $null
$frontendProc = $null

try {
    Stop-Port $backendPort
    Stop-Port $frontendPort

    if ($target -eq "backend" -or $target -eq "both") {
        $backendProc = Start-Backend
    }

    if ($target -eq "frontend" -or $target -eq "both") {
        $frontendProc = Start-Frontend
    }

    Write-Host ""
    Write-Host "=== System ready ===" -ForegroundColor Green
    Write-Host "  Backend  : http://localhost:$backendPort" -ForegroundColor White
    Write-Host "  Frontend : http://localhost:$frontendPort" -ForegroundColor White
    Write-Host "  Login    : admin / admin123" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Press Ctrl+C to stop all services" -ForegroundColor Yellow
    Write-Host ""

    # Wait forever until Ctrl+C
    while ($true) {
        Start-Sleep -Seconds 1
    }
}
finally {
    Write-Host ""
    Write-Host "=== Stopping services ===" -ForegroundColor Yellow
    if ($backendProc) {
        $backendProc.Kill()
        Write-Host "  Backend stopped" -ForegroundColor Gray
    }
    if ($frontendProc) {
        $frontendProc.Kill()
        Write-Host "  Frontend stopped" -ForegroundColor Gray
    }
    Stop-Port $backendPort
    Stop-Port $frontendPort
    Write-Host "  All services stopped." -ForegroundColor Green
}
