#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Baseline-Checks for NetBird Machine Tunnel E2E Tests (T-6.7)
.DESCRIPTION
    Verifies system baseline before running E2E tests:
    - Zeit-Synchronisation (Stratum)
    - DNS-Konfiguration
    - Routing-Tabelle
    - Netzwerk-Interfaces
    - Kritische Dienste (W32Time, Dnscache, NlaSvc)
    - WireGuard Tunnel Status
    - DC Connectivity
    - Kerberos TGT (SYSTEM Session)
.EXAMPLE
    .\Test-Baseline.ps1
.NOTES
    Author: NetBird Machine Tunnel Fork
    Version: 1.0.0
#>

Write-Host "=== T-6.7: Baseline-Checks ===" -ForegroundColor Cyan
Write-Host "Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')`n"

$allPass = $true

# 1. Zeit-Synchronisation
Write-Host "1. Zeit-Synchronisation..." -ForegroundColor Yellow
$timeStatus = w32tm /query /status 2>&1 | Out-String
$stratum = if ($timeStatus -match "Stratum:\s*(\d+)") { $matches[1] } else { "N/A" }
Write-Host "   Stratum: $stratum"
if ([int]$stratum -le 3) {
    Write-Host "   [PASS] Zeit-Synchronisation OK" -ForegroundColor Green
} else {
    Write-Host "   [WARN] Stratum hoch" -ForegroundColor Yellow
}

# 2. DNS-Konfiguration
Write-Host "`n2. DNS-Konfiguration..." -ForegroundColor Yellow
$dnsServers = Get-DnsClientServerAddress -AddressFamily IPv4 | Where-Object { $_.ServerAddresses }
foreach ($dns in $dnsServers) {
    Write-Host "   $($dns.InterfaceAlias) - $($dns.ServerAddresses -join ', ')"
}

# 3. Routing
Write-Host "`n3. Routing..." -ForegroundColor Yellow
$defaultGw = (Get-NetRoute -DestinationPrefix "0.0.0.0/0" -ErrorAction SilentlyContinue).NextHop | Select-Object -First 1
Write-Host "   Default Gateway: $defaultGw"

$dcRoute = Get-NetRoute -DestinationPrefix "192.168.100.0/24" -ErrorAction SilentlyContinue
if ($dcRoute) {
    Write-Host "   [INFO] Route zu DC-Netz via $($dcRoute.NextHop)" -ForegroundColor Gray
} else {
    Write-Host "   [INFO] Keine Route zu DC-Netz" -ForegroundColor Gray
}

# 4. Interface-Status
Write-Host "`n4. Netzwerk-Interfaces..." -ForegroundColor Yellow
$adapters = Get-NetAdapter | Where-Object { $_.Status -eq "Up" } | Select-Object Name, Status
foreach ($a in $adapters) {
    Write-Host "   $($a.Name) - $($a.Status)"
}

# 5. Kritische Dienste
Write-Host "`n5. Kritische Dienste..." -ForegroundColor Yellow
$services = @("W32Time", "Dnscache", "NlaSvc", "NetBirdMachine")
foreach ($svcName in $services) {
    $s = Get-Service -Name $svcName -ErrorAction SilentlyContinue
    if ($s) {
        $color = if ($s.Status -eq "Running") { "Green" } else { "Red" }
        Write-Host "   $($s.Name) - $($s.Status)" -ForegroundColor $color
        if ($s.Status -ne "Running" -and $svcName -ne "NetBirdMachine") { $allPass = $false }
    } else {
        Write-Host "   $svcName - Not found" -ForegroundColor Yellow
    }
}

# 6. WireGuard Interface
Write-Host "`n6. WireGuard Tunnel..." -ForegroundColor Yellow
$wg = Get-NetAdapter -InterfaceDescription "WireGuard*" -ErrorAction SilentlyContinue
if ($wg -and $wg.Status -eq "Up") {
    $wgIP = (Get-NetIPAddress -InterfaceIndex $wg.ifIndex -AddressFamily IPv4 -ErrorAction SilentlyContinue).IPAddress
    Write-Host "   [PASS] Interface $($wg.Name), IP $wgIP" -ForegroundColor Green
} else {
    Write-Host "   [FAIL] WireGuard interface not up" -ForegroundColor Red
    $allPass = $false
}

# 7. DC Connectivity
Write-Host "`n7. DC Connectivity..." -ForegroundColor Yellow
$dc = Test-NetConnection -ComputerName 192.168.100.20 -Port 389 -WarningAction SilentlyContinue
if ($dc.TcpTestSucceeded) {
    Write-Host "   [PASS] DC LDAP (389) erreichbar" -ForegroundColor Green
} else {
    Write-Host "   [FAIL] DC LDAP nicht erreichbar" -ForegroundColor Red
    $allPass = $false
}

# 8. Kerberos TGT
Write-Host "`n8. Kerberos TGT (SYSTEM)..." -ForegroundColor Yellow
$klist = klist -li 0x3e7 2>&1 | Out-String
if ($klist -match "krbtgt/" -and $klist -match "Cached Tickets:") {
    $ticketCount = ([regex]::Matches($klist, "#\d+>")).Count
    Write-Host "   [PASS] Machine TGT found ($ticketCount tickets)" -ForegroundColor Green
} else {
    Write-Host "   [FAIL] No Machine TGT" -ForegroundColor Red
    $allPass = $false
}

# Summary
Write-Host "`n=== BASELINE SUMMARY ===" -ForegroundColor Cyan
if ($allPass) {
    Write-Host "[PASS] All baseline checks passed" -ForegroundColor Green
    exit 0
} else {
    Write-Host "[FAIL] Some baseline checks failed" -ForegroundColor Red
    exit 1
}
