# Lab Management System — Reliable Dev Launcher
# Usage: .\dev.ps1

param([string]$target = "both")

$ErrorActionPreference = "Stop"
$backendDir = "$PSScriptRoot\backend"
$frontendDir = "$PSScriptRoot\frontend"
$backendPort = 8080
$frontendPort = 5173

# ── Kill anything on dev ports ───────────────────────────────────
Write-Host "[kill] Clearing ports $backendPort / $frontendPort..." -ForegroundColor Yellow

# taskkill can fail when process already exited; don't let that stop us.
$prevEAP = $ErrorActionPreference
$ErrorActionPreference = "Continue"

cmd /c "for /f `"tokens=5`" %a in ('netstat -ano ^| findstr `":$backendPort `" ^| findstr LISTENING') do taskkill /F /PID %a" 2>$null
cmd /c "for /f `"tokens=5`" %a in ('netstat -ano ^| findstr `":$frontendPort `" ^| findstr LISTENING') do taskkill /F /PID %a" 2>$null

$ErrorActionPreference = $prevEAP
Start-Sleep -Seconds 2

# Verify ports are free
$b = netstat -ano | Select-String ":$backendPort " | Select-String "LISTENING"
$f = netstat -ano | Select-String ":$frontendPort " | Select-String "LISTENING"
if ($b) { Write-Host "  WARNING: Port $backendPort still occupied!" -ForegroundColor Red }
if ($f) { Write-Host "  WARNING: Port $frontendPort still occupied!" -ForegroundColor Red }
if (-not $b -and -not $f) { Write-Host "  Ports cleared." -ForegroundColor Green }

# ── Helper: Wait for a port to be listening ──────────────────────
function Wait-Port($port, $timeoutSec = 20) {
    $elapsed = 0
    while ($elapsed -lt $timeoutSec) {
        $listening = netstat -ano | Select-String ":$port " | Select-String "LISTENING"
        if ($listening) { return $true }
        Start-Sleep -Seconds 1
        $elapsed++
    }
    return $false
}

# ── Start Backend ────────────────────────────────────────────────
function Start-Backend {
    Write-Host ""
    Write-Host "=== [1/2] Starting backend ===" -ForegroundColor Cyan
    Write-Host "  URL: http://localhost:$backendPort" -ForegroundColor Gray

    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = (Get-Command go -ErrorAction Stop).Source
    $psi.Arguments = "run ./cmd/main.go"
    $psi.WorkingDirectory = $backendDir
    $psi.UseShellExecute = $false
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true
    $psi.Environment["PATH"] = $env:PATH

    $proc = [System.Diagnostics.Process]::Start($psi)

    Write-Host "  Waiting for backend..." -ForegroundColor Gray
    $ready = Wait-Port $backendPort 20
    if (-not $ready) {
        $stderr = $proc.StandardError.ReadToEnd()
        Write-Host "  ERROR: Backend failed to start within 20s" -ForegroundColor Red
        if ($stderr) { Write-Host "  $stderr" -ForegroundColor DarkRed }
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

    if (-not (Test-Path "$frontendDir\node_modules")) {
        Write-Host "  Installing dependencies..." -ForegroundColor Yellow
        & npm --prefix $frontendDir install
    }

    # Only clear Vite cache if explicitly broken (not every start)
    # Vite 6.x pre-bundles deps on cold start; clearing cache forces re-bundle
    # which can exceed the 15s timeout on slower machines.
    # $viteCache = "$frontendDir\node_modules\.vite"
    # if (Test-Path $viteCache) {
    #     Remove-Item -Recurse -Force $viteCache
    #     Write-Host "  Vite cache cleared" -ForegroundColor DarkGray
    # }

    $psi = New-Object System.Diagnostics.ProcessStartInfo
    $psi.FileName = "cmd.exe"
    $psi.Arguments = "/c `"cd /d $frontendDir && npx vite`""
    $psi.WorkingDirectory = $frontendDir
    $psi.UseShellExecute = $false
    $psi.RedirectStandardOutput = $true
    $psi.RedirectStandardError = $true

    $proc = [System.Diagnostics.Process]::Start($psi)

    Write-Host "  Waiting for frontend (may take up to 30s on first start)..." -ForegroundColor Gray
    $ready = Wait-Port $frontendPort 30
    if (-not $ready) {
        $stderr = $proc.StandardError.ReadToEnd()
        Write-Host "  ERROR: Frontend failed to start within 30s" -ForegroundColor Red
        if ($stderr) { Write-Host "  $stderr" -ForegroundColor DarkRed }
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
    Write-Host "  IMPORTANT: In browser, press Ctrl+Shift+R (hard refresh)" -ForegroundColor Yellow
    Write-Host "             If login still fails: F12 → Application → Clear site data" -ForegroundColor Yellow
    Write-Host ""

    while ($true) { Start-Sleep -Seconds 1 }
}
finally {
    Write-Host ""
    Write-Host "=== Stopping services ===" -ForegroundColor Yellow
    if ($backendProc) { $backendProc.Kill() }
    if ($frontendProc) { $frontendProc.Kill() }
    $prevEAP = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    cmd /c "for /f `"tokens=5`" %a in ('netstat -ano ^| findstr `":$backendPort `" ^| findstr LISTENING') do taskkill /F /PID %a" 2>$null
    cmd /c "for /f `"tokens=5`" %a in ('netstat -ano ^| findstr `":$frontendPort `" ^| findstr LISTENING') do taskkill /F /PID %a" 2>$null
    $ErrorActionPreference = $prevEAP
    Write-Host "  All services stopped." -ForegroundColor Green
}
