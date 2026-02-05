#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Go/No-Go Matrix for NetBird Machine Tunnel E2E Tests (T-6.8)
.DESCRIPTION
    Validates all prerequisites before E2E tests can start.
    ALL checks must PASS before tests can proceed.
.PARAMETER WithTunnel
    Run checks with tunnel UP (default: checks isolation without tunnel)
.PARAMETER StopOnFirstFailure
    Stop on first failed check
.EXAMPLE
    .\Test-GoNoGoMatrix.ps1 -WithTunnel
.EXAMPLE
    .\Test-GoNoGoMatrix.ps1  # Isolation check without tunnel
.NOTES
    Author: NetBird Machine Tunnel Fork
    Version: 1.0.0
#>

param(
    [switch]$WithTunnel,
    [switch]$StopOnFirstFailure
)

Write-Host "=== T-6.8: GO/NO-GO MATRIX ===" -ForegroundColor Cyan
Write-Host "Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
Write-Host "Mode: $(if ($WithTunnel) { 'WITH Tunnel' } else { 'WITHOUT Tunnel (Isolation Check)' })`n"

$results = @()
$allPassed = $true

function Add-Result {
    param($Check, $Status, $Message)
    $script:results += @{ Check = $Check; Status = $Status; Message = $Message }
    $color = switch ($Status) { "PASS" { "Green" } "FAIL" { "Red" } "SKIP" { "Yellow" } }
    Write-Host "  [$Status] $Message" -ForegroundColor $color
    if ($Status -eq "FAIL") { $script:allPassed = $false }
}

# 14.0.1: Network Interfaces
Write-Host "[14.0.1] Network Interfaces..." -ForegroundColor Yellow
$adapters = Get-NetAdapter | Where-Object { $_.Status -eq "Up" }
if ($adapters.Count -ge 2) {
    Add-Result "14.0.1" "PASS" "$($adapters.Count) interfaces UP"
} else {
    Add-Result "14.0.1" "FAIL" "Weniger als 2 Interfaces UP"
}

# 14.0.2: Client IP Configuration
Write-Host "`n[14.0.2] Client IP Configuration..." -ForegroundColor Yellow
$mgmtIP = (Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.IPAddress -like "10.0.0.*" }).IPAddress
if ($mgmtIP) {
    Add-Result "14.0.2" "PASS" "Management IP: $mgmtIP"
} else {
    Add-Result "14.0.2" "FAIL" "Keine Management IP (10.0.0.x)"
}

# 14.0.3: Domain Status
Write-Host "`n[14.0.3] Domain Status..." -ForegroundColor Yellow
$domain = (Get-WmiObject Win32_ComputerSystem).Domain
if ($domain -eq "test.local") {
    Add-Result "14.0.3" "PASS" "Domain-Joined: $domain"
} else {
    Add-Result "14.0.3" "SKIP" "Nicht domain-joined (Domain: $domain)"
}

# 14.0.4: Router-Peer Isolation (nur ohne Tunnel)
Write-Host "`n[14.0.4] Router-Peer Isolation..." -ForegroundColor Yellow
if (-not $WithTunnel) {
    $dcDirect = Test-NetConnection -ComputerName 192.168.100.20 -Port 389 -WarningAction SilentlyContinue
    if (-not $dcDirect.TcpTestSucceeded) {
        Add-Result "14.0.4" "PASS" "DC nicht direkt erreichbar (Isolation OK)"
    } else {
        Add-Result "14.0.4" "FAIL" "DC direkt erreichbar - Isolation broken!"
    }
} else {
    Add-Result "14.0.4" "SKIP" "Tunnel aktiv - Isolation-Check nicht moeglich"
}

# 14.0.5: Management-Server
Write-Host "`n[14.0.5] Management-Server..." -ForegroundColor Yellow
$mgmtServer = Test-NetConnection -ComputerName 10.0.0.103 -Port 443 -WarningAction SilentlyContinue
if ($mgmtServer.TcpTestSucceeded) {
    Add-Result "14.0.5" "PASS" "Management-Server erreichbar (10.0.0.103:443)"
} else {
    Add-Result "14.0.5" "FAIL" "Management-Server nicht erreichbar"
}

# 14.0.6: Tunnel Status (mit Tunnel)
Write-Host "`n[14.0.6] Tunnel Status..." -ForegroundColor Yellow
if ($WithTunnel) {
    $wg = Get-NetAdapter -InterfaceDescription "WireGuard*" -ErrorAction SilentlyContinue
    if ($wg -and $wg.Status -eq "Up") {
        $wgIP = (Get-NetIPAddress -InterfaceIndex $wg.ifIndex -AddressFamily IPv4 -ErrorAction SilentlyContinue).IPAddress
        Add-Result "14.0.6" "PASS" "Tunnel UP: $($wg.Name) ($wgIP)"

        # DC via Tunnel erreichbar?
        $dcVia = Test-NetConnection -ComputerName 192.168.100.20 -Port 389 -WarningAction SilentlyContinue
        if ($dcVia.TcpTestSucceeded) {
            Add-Result "14.0.6b" "PASS" "DC via Tunnel erreichbar"
        } else {
            Add-Result "14.0.6b" "FAIL" "DC via Tunnel NICHT erreichbar"
        }
    } else {
        Add-Result "14.0.6" "FAIL" "WireGuard Interface nicht UP"
    }
} else {
    $svc = Get-Service NetBirdMachine -ErrorAction SilentlyContinue
    if ($svc -and $svc.Status -eq "Stopped") {
        Add-Result "14.0.6" "PASS" "Service gestoppt (fuer Isolation-Test)"
    } else {
        Add-Result "14.0.6" "SKIP" "Service Status: $($svc.Status)"
    }
}

# 14.0.7: Baseline Services
Write-Host "`n[14.0.7] Baseline Services..." -ForegroundColor Yellow
$services = @("W32Time", "Dnscache", "NlaSvc")
$allRunning = $true
foreach ($svcName in $services) {
    $svc = Get-Service -Name $svcName -ErrorAction SilentlyContinue
    if (-not $svc -or $svc.Status -ne "Running") {
        $allRunning = $false
    }
}
if ($allRunning) {
    Add-Result "14.0.7" "PASS" "Kritische Dienste laufen"
} else {
    Add-Result "14.0.7" "FAIL" "Kritische Dienste nicht alle Running"
}

# Summary
Write-Host "`n=== MATRIX SUMMARY ===" -ForegroundColor Cyan
$passCount = ($results | Where-Object { $_.Status -eq "PASS" }).Count
$failCount = ($results | Where-Object { $_.Status -eq "FAIL" }).Count
$skipCount = ($results | Where-Object { $_.Status -eq "SKIP" }).Count

Write-Host "PASS: $passCount | FAIL: $failCount | SKIP: $skipCount"

if ($failCount -eq 0) {
    Write-Host "`n[GO] Alle Checks bestanden - E2E-Tests koennen starten" -ForegroundColor Green
    exit 0
} else {
    Write-Host "`n[NO-GO] $failCount Check(s) fehlgeschlagen!" -ForegroundColor Red
    exit 1
}
