#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Verifies that all NetBird NRPT rules are removed.

.DESCRIPTION
    Checks both Registry paths AND PowerShell cmdlet view for completeness.
    v3.6: Registry-based verification (consistent with service implementation)

    Checks:
    1. Service Registry path (Dnscache\Parameters\DnsPolicyConfig)
    2. Policy Registry path (Windows NT\DNSClient\DnsPolicyConfig)
    3. PowerShell cmdlet view (Get-DnsClientNrptRule)

.EXAMPLE
    .\verify-nrpt-cleanup.ps1

.NOTES
    Author: NetBird Machine Tunnel Fork (Test-Lab Scripts)
    Purpose: Development/Testing - verify cleanup after reset
    Version: 1.0.0
#>

[CmdletBinding()]
param()

Set-StrictMode -Version Latest

Write-Host ""
Write-Host "=== NetBird NRPT Cleanup Verification ===" -ForegroundColor Cyan
Write-Host ""

$allClean = $true

# Check 1: Registry (Service-Pfad)
Write-Host "Check 1: Service Registry Path" -ForegroundColor Yellow
$servicePath = "HKLM:\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig"
if (Test-Path $servicePath) {
    $serviceRules = Get-ChildItem -Path $servicePath -ErrorAction SilentlyContinue |
        Where-Object { $_.PSChildName -like "NetBird-Machine-*" }
    if ($serviceRules) {
        Write-Host "  [FAIL] Found $($serviceRules.Count) NetBird rule(s):" -ForegroundColor Red
        $serviceRules | ForEach-Object { Write-Host "    - $($_.PSChildName)" -ForegroundColor Gray }
        $allClean = $false
    } else {
        Write-Host "  [OK] Clean" -ForegroundColor Green
    }
} else {
    Write-Host "  [OK] Path does not exist" -ForegroundColor Green
}

# Check 2: Registry (Policy-Pfad)
Write-Host "`nCheck 2: Policy Registry Path" -ForegroundColor Yellow
$policyPath = "HKLM:\SOFTWARE\Policies\Microsoft\Windows NT\DNSClient\DnsPolicyConfig"
if (Test-Path $policyPath) {
    $policyRules = Get-ChildItem -Path $policyPath -ErrorAction SilentlyContinue |
        Where-Object { $_.PSChildName -like "NetBird-Machine-*" }
    if ($policyRules) {
        Write-Host "  [FAIL] Found $($policyRules.Count) NetBird rule(s):" -ForegroundColor Red
        $policyRules | ForEach-Object { Write-Host "    - $($_.PSChildName)" -ForegroundColor Gray }
        $allClean = $false
    } else {
        Write-Host "  [OK] Clean" -ForegroundColor Green
    }
} else {
    Write-Host "  [OK] Path does not exist" -ForegroundColor Green
}

# Check 3: PowerShell Cmdlet View
Write-Host "`nCheck 3: PowerShell Cmdlet View" -ForegroundColor Yellow
try {
    $psRules = Get-DnsClientNrptRule -ErrorAction SilentlyContinue |
        Where-Object { $_.Comment -like "*NetBird*" -or $_.Name -like "*NetBird*" }
    if ($psRules) {
        Write-Host "  [FAIL] Found $($psRules.Count) NetBird rule(s):" -ForegroundColor Red
        $psRules | ForEach-Object {
            Write-Host "    - $($_.Name): $($_.Namespace -join ', ')" -ForegroundColor Gray
        }
        $allClean = $false
    } else {
        Write-Host "  [OK] Clean" -ForegroundColor Green
    }
} catch {
    Write-Host "  [WARN] Could not query NRPT rules: $_" -ForegroundColor Yellow
}

# Check 4: Firewall Rules
Write-Host "`nCheck 4: Firewall Rules" -ForegroundColor Yellow
$fwRules = Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue
$fwRulesLegacy = Get-NetFirewallRule -Group "NetBird Machine" -ErrorAction SilentlyContinue
$totalFw = @($fwRules).Count + @($fwRulesLegacy).Count
if ($totalFw -gt 0) {
    Write-Host "  [FAIL] Found $totalFw firewall rule(s)" -ForegroundColor Red
    $allClean = $false
} else {
    Write-Host "  [OK] Clean" -ForegroundColor Green
}

# Check 5: WireGuard Interface
Write-Host "`nCheck 5: WireGuard Interface" -ForegroundColor Yellow
$adapter = Get-NetAdapter -Name "wg-nb-machine" -ErrorAction SilentlyContinue
if ($adapter) {
    Write-Host "  [FAIL] Interface still exists: wg-nb-machine" -ForegroundColor Red
    $allClean = $false
} else {
    Write-Host "  [OK] Clean" -ForegroundColor Green
}

# Check 6: Service
Write-Host "`nCheck 6: Service Registration" -ForegroundColor Yellow
$service = Get-Service -Name "NetBirdMachine" -ErrorAction SilentlyContinue
if ($service) {
    Write-Host "  [FAIL] Service still registered: NetBirdMachine ($($service.Status))" -ForegroundColor Red
    $allClean = $false
} else {
    Write-Host "  [OK] Clean" -ForegroundColor Green
}

# Check 7: Config Directory
Write-Host "`nCheck 7: Config Directory" -ForegroundColor Yellow
$configPath = "C:\ProgramData\NetBird"
if (Test-Path $configPath) {
    $files = Get-ChildItem -Path $configPath -Recurse -ErrorAction SilentlyContinue
    Write-Host "  [FAIL] Config directory exists with $($files.Count) file(s)" -ForegroundColor Red
    $allClean = $false
} else {
    Write-Host "  [OK] Clean" -ForegroundColor Green
}

# Summary
Write-Host ""
Write-Host "=============================================" -ForegroundColor Cyan
if ($allClean) {
    Write-Host " All checks passed - system is clean!" -ForegroundColor Green
    exit 0
} else {
    Write-Host " Cleanup incomplete - run reset-netbird-machine.ps1" -ForegroundColor Red
    exit 1
}
