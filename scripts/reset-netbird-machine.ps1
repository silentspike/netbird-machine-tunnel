#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Removes all NetBird Machine Tunnel artifacts for clean re-installation.

.DESCRIPTION
    This script removes all NetBird Machine Tunnel components:
    - Windows Service (NetBirdMachine)
    - WireGuard interface (wg-nb-machine)
    - NRPT rules (Registry-based, both paths)
    - Firewall rules (NetBird Machine Tunnel group)
    - Configuration directory (C:\ProgramData\NetBird)
    - Optionally: Machine certificates

    v3.6: Registry-based NRPT removal (consistent with service implementation)

.PARAMETER Force
    Skip confirmation prompts.

.PARAMETER RemoveCerts
    Also remove machine certificates from LocalMachine\My store.

.PARAMETER WhatIf
    Show what would be removed without actually removing anything.

.EXAMPLE
    .\reset-netbird-machine.ps1
    # Interactive mode with confirmation prompts

.EXAMPLE
    .\reset-netbird-machine.ps1 -Force
    # Non-interactive mode, no prompts

.EXAMPLE
    .\reset-netbird-machine.ps1 -WhatIf
    # Dry-run mode, shows what would be removed

.NOTES
    Author: NetBird Machine Tunnel Fork (Test-Lab Scripts)
    Purpose: Development/Testing - clean reset between test runs
    Version: 1.0.0
#>

[CmdletBinding(SupportsShouldProcess)]
param(
    [switch]$Force,
    [switch]$RemoveCerts
)

# Strict mode
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "=== NetBird Machine Tunnel Reset ===" -ForegroundColor Cyan
Write-Host ""

if ($WhatIfPreference) {
    Write-Host "[DRY-RUN MODE] No changes will be made." -ForegroundColor Yellow
    Write-Host ""
}

# Confirmation prompt (unless -Force or -WhatIf)
if (-not $Force -and -not $WhatIfPreference) {
    Write-Host "This will remove ALL NetBird Machine Tunnel artifacts:" -ForegroundColor Yellow
    Write-Host "  - Service (NetBirdMachine)"
    Write-Host "  - WireGuard interface (wg-nb-machine)"
    Write-Host "  - NRPT rules"
    Write-Host "  - Firewall rules"
    Write-Host "  - Config directory (C:\ProgramData\NetBird)"
    if ($RemoveCerts) {
        Write-Host "  - Machine certificates" -ForegroundColor Red
    }
    Write-Host ""
    $confirm = Read-Host "Continue? (y/N)"
    if ($confirm -ne "y" -and $confirm -ne "Y") {
        Write-Host "Aborted." -ForegroundColor Gray
        exit 0
    }
    Write-Host ""
}

# Step 1: Stop Service
Write-Host "Step 1: Stopping service..." -ForegroundColor Yellow
$service = Get-Service -Name "NetBirdMachine" -ErrorAction SilentlyContinue
if ($service) {
    if ($service.Status -eq "Running") {
        if ($PSCmdlet.ShouldProcess("NetBirdMachine", "Stop service")) {
            Stop-Service -Name "NetBirdMachine" -Force -ErrorAction SilentlyContinue
            Start-Sleep -Seconds 2
            Write-Host "  Service stopped" -ForegroundColor Green
        }
    } else {
        Write-Host "  Service already stopped" -ForegroundColor Gray
    }
} else {
    Write-Host "  Service not found" -ForegroundColor Gray
}

# Step 2: Remove WireGuard Interface (by name, consistent with service)
Write-Host "`nStep 2: Removing WireGuard interface..." -ForegroundColor Yellow
$interfaceName = "wg-nb-machine"
$adapter = Get-NetAdapter -Name $interfaceName -ErrorAction SilentlyContinue
if ($adapter) {
    if ($PSCmdlet.ShouldProcess($interfaceName, "Remove interface")) {
        # Remove IP addresses first
        Remove-NetIPAddress -InterfaceAlias $interfaceName -Confirm:$false -ErrorAction SilentlyContinue
        # Disable adapter (triggers cleanup)
        netsh interface set interface $interfaceName admin=disable 2>&1 | Out-Null
        Start-Sleep -Milliseconds 500
        Write-Host "  Interface removed: $interfaceName (GUID: $($adapter.InterfaceGuid))" -ForegroundColor Green
    }
} else {
    Write-Host "  No WireGuard interface found" -ForegroundColor Gray
}

# Step 3: Remove NRPT Rules (Registry-based - v3.6)
Write-Host "`nStep 3: Removing NRPT rules (Registry)..." -ForegroundColor Yellow
$nrptPaths = @(
    "HKLM:\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig",
    "HKLM:\SOFTWARE\Policies\Microsoft\Windows NT\DNSClient\DnsPolicyConfig"
)

$removedNrpt = 0
foreach ($basePath in $nrptPaths) {
    if (-not (Test-Path $basePath)) { continue }

    $rules = Get-ChildItem -Path $basePath -ErrorAction SilentlyContinue
    foreach ($rule in $rules) {
        if ($rule.PSChildName -like "NetBird-Machine-*") {
            if ($PSCmdlet.ShouldProcess($rule.PSChildName, "Remove NRPT rule")) {
                Remove-Item -Path $rule.PSPath -Recurse -Force
                Write-Host "    Removed: $($rule.PSChildName)" -ForegroundColor Gray
                $removedNrpt++
            }
        }
    }
}

if ($removedNrpt -gt 0) {
    Write-Host "  Removed $removedNrpt NRPT rule(s)" -ForegroundColor Green
    if ($PSCmdlet.ShouldProcess("DNS Cache", "Flush")) {
        ipconfig /flushdns 2>&1 | Out-Null
        Write-Host "  DNS cache flushed" -ForegroundColor Green
    }
} else {
    Write-Host "  No NetBird NRPT rules found" -ForegroundColor Gray
}

# Step 4: Remove Firewall Rules
Write-Host "`nStep 4: Removing firewall rules..." -ForegroundColor Yellow
$fwRules = Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue
if ($fwRules) {
    if ($PSCmdlet.ShouldProcess("NetBird Machine Tunnel", "Remove firewall rules")) {
        $fwRules | Remove-NetFirewallRule -ErrorAction SilentlyContinue
        Write-Host "  Removed $($fwRules.Count) firewall rule(s)" -ForegroundColor Green
    }
} else {
    # Try legacy group name
    $fwRulesLegacy = Get-NetFirewallRule -Group "NetBird Machine" -ErrorAction SilentlyContinue
    if ($fwRulesLegacy) {
        if ($PSCmdlet.ShouldProcess("NetBird Machine", "Remove firewall rules (legacy)")) {
            $fwRulesLegacy | Remove-NetFirewallRule -ErrorAction SilentlyContinue
            Write-Host "  Removed $($fwRulesLegacy.Count) firewall rule(s) (legacy)" -ForegroundColor Green
        }
    } else {
        Write-Host "  No NetBird firewall rules found" -ForegroundColor Gray
    }
}

# Step 5: Remove Service Registration
Write-Host "`nStep 5: Removing service registration..." -ForegroundColor Yellow
$service = Get-Service -Name "NetBirdMachine" -ErrorAction SilentlyContinue
if ($service) {
    if ($PSCmdlet.ShouldProcess("NetBirdMachine", "Delete service")) {
        sc.exe delete NetBirdMachine 2>&1 | Out-Null
        Write-Host "  Service registration removed" -ForegroundColor Green
    }
} else {
    Write-Host "  Service not registered" -ForegroundColor Gray
}

# Step 6: Kill any remaining NetBird processes
Write-Host "`nStep 6: Killing remaining processes..." -ForegroundColor Yellow
$procs = @(Get-Process -Name "netbird*" -ErrorAction SilentlyContinue)
if ($procs.Count -gt 0) {
    if ($PSCmdlet.ShouldProcess("netbird processes", "Kill")) {
        $procs | Stop-Process -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
        Write-Host "  Killed $($procs.Count) process(es)" -ForegroundColor Green
    }
} else {
    Write-Host "  No NetBird processes running" -ForegroundColor Gray
}

# Step 7: Remove Config Directory
Write-Host "`nStep 7: Removing config directory..." -ForegroundColor Yellow
$configPath = "C:\ProgramData\NetBird"
if (Test-Path $configPath) {
    if ($PSCmdlet.ShouldProcess($configPath, "Remove directory")) {
        # Retry with delay if file is locked
        $retries = 3
        for ($i = 0; $i -lt $retries; $i++) {
            try {
                Remove-Item -Path $configPath -Recurse -Force -ErrorAction Stop
                Write-Host "  Config removed: $configPath" -ForegroundColor Green
                break
            } catch {
                if ($i -lt $retries - 1) {
                    Write-Host "  Retrying in 2 seconds..." -ForegroundColor Gray
                    Start-Sleep -Seconds 2
                } else {
                    Write-Host "  Failed to remove config: $_" -ForegroundColor Red
                }
            }
        }
    }
} else {
    Write-Host "  No config directory found" -ForegroundColor Gray
}

# Step 8: Optional - Remove Certificates
if ($RemoveCerts) {
    Write-Host "`nStep 8: Removing machine certificates..." -ForegroundColor Yellow
    $certs = Get-ChildItem Cert:\LocalMachine\My -ErrorAction SilentlyContinue | Where-Object {
        # Match certificates with NetBird-related attributes
        $_.Subject -like "*NetBird*" -or
        $_.FriendlyName -like "*NetBird*" -or
        ($_.Extensions | Where-Object { $_.Oid.FriendlyName -eq "Certificate Template Name" })
    }

    if ($certs) {
        foreach ($cert in $certs) {
            if ($PSCmdlet.ShouldProcess($cert.Subject, "Remove certificate")) {
                Remove-Item $cert.PSPath -Force
                Write-Host "    Removed: $($cert.Subject)" -ForegroundColor Gray
            }
        }
        Write-Host "  Removed $($certs.Count) certificate(s)" -ForegroundColor Green
    } else {
        Write-Host "  No NetBird certificates found" -ForegroundColor Gray
    }
}

Write-Host ""
Write-Host "=== Reset Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Run verify-nrpt-cleanup.ps1 to confirm cleanup"
Write-Host "  2. Reinstall: .\netbird-machine.exe install"
Write-Host ""
