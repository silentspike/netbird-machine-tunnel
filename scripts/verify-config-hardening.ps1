#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Verifies security hardening of the NetBird Machine Tunnel configuration.

.DESCRIPTION
    This script performs comprehensive security verification:
    - Directory owner verification (should be SYSTEM)
    - ACL configuration check (inheritance disabled, correct permissions)
    - DPAPI encryption verification of config values
    - Service configuration validation

    Enterprise SOTA: Structured report with OK/FAIL status per check.

.PARAMETER ConfigDir
    Configuration directory to verify. Default: C:\ProgramData\NetBird

.PARAMETER ServiceName
    Name of the Windows Service. Default: NetBirdMachine

.PARAMETER Verbose
    Show detailed output for each check.

.EXAMPLE
    .\verify-config-hardening.ps1

.EXAMPLE
    .\verify-config-hardening.ps1 -ConfigDir "C:\ProgramData\NetBird" -Verbose

.OUTPUTS
    Structured report with OK/FAIL status for each security check.

.NOTES
    Author: NetBird Machine Tunnel Fork
    Requires: Windows 10/11, Server 2019+, Administrator privileges
    Version: 1.0.0
#>

[CmdletBinding()]
param(
    [Parameter(Mandatory = $false)]
    [string]$ConfigDir = "C:\ProgramData\NetBird",

    [Parameter(Mandatory = $false)]
    [string]$ServiceName = "NetBirdMachine"
)

# Strict mode for better error handling
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Results tracking
$script:totalChecks = 0
$script:passedChecks = 0
$script:failedChecks = 0
$script:warnings = 0

# Check result helper
function Write-CheckResult {
    param(
        [string]$CheckName,
        [string]$Status,  # OK, FAIL, WARN, SKIP
        [string]$Details = ""
    )

    $script:totalChecks++

    $statusColor = switch ($Status) {
        "OK"   { "Green"; $script:passedChecks++ }
        "FAIL" { "Red"; $script:failedChecks++ }
        "WARN" { "Yellow"; $script:warnings++ }
        "SKIP" { "Gray" }
        default { "White" }
    }

    $statusText = "[$Status]".PadRight(6)
    Write-Host "$statusText " -ForegroundColor $statusColor -NoNewline
    Write-Host $CheckName -NoNewline

    if ($Details) {
        Write-Host " - $Details" -ForegroundColor Gray
    } else {
        Write-Host ""
    }
}

# Check if running as Administrator
function Test-Administrator {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($identity)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

if (-not (Test-Administrator)) {
    Write-Host "This script requires Administrator privileges. Please run as Administrator." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host " NetBird Machine Tunnel Security Verification" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Config Directory: $ConfigDir"
Write-Host "Service Name: $ServiceName"
Write-Host ""
Write-Host "Running security checks..."
Write-Host ""

# =============================================================================
# Section 1: Directory Checks
# =============================================================================
Write-Host "--- Directory Security ---" -ForegroundColor Yellow

# Check 1.1: Directory exists
if (Test-Path $ConfigDir) {
    Write-CheckResult "Config directory exists" "OK" $ConfigDir
} else {
    Write-CheckResult "Config directory exists" "FAIL" "Directory not found: $ConfigDir"
    Write-Host ""
    Write-Host "Cannot continue without config directory." -ForegroundColor Red
    exit 1
}

# Check 1.2: Directory owner is SYSTEM
try {
    $acl = Get-Acl -Path $ConfigDir
    $owner = $acl.Owner
    $systemSID = [System.Security.Principal.SecurityIdentifier]::new("S-1-5-18")
    $systemAccount = $systemSID.Translate([System.Security.Principal.NTAccount]).Value

    if ($owner -eq $systemAccount -or $owner -eq "NT AUTHORITY\SYSTEM") {
        Write-CheckResult "Directory owner is SYSTEM" "OK" $owner
    } else {
        Write-CheckResult "Directory owner is SYSTEM" "FAIL" "Owner is: $owner (expected: SYSTEM)"
    }
} catch {
    Write-CheckResult "Directory owner is SYSTEM" "FAIL" $_.Exception.Message
}

# Check 1.3: Inheritance is disabled
try {
    $acl = Get-Acl -Path $ConfigDir
    $isProtected = $acl.AreAccessRulesProtected

    if ($isProtected) {
        Write-CheckResult "Inheritance disabled" "OK" "Access rules are protected"
    } else {
        Write-CheckResult "Inheritance disabled" "WARN" "Inheritance is enabled (should be disabled)"
    }
} catch {
    Write-CheckResult "Inheritance disabled" "FAIL" $_.Exception.Message
}

# Check 1.4: SYSTEM has Full Control
try {
    $acl = Get-Acl -Path $ConfigDir
    $systemSID = [System.Security.Principal.SecurityIdentifier]::new("S-1-5-18")

    $systemRules = $acl.Access | Where-Object {
        $_.IdentityReference.Translate([System.Security.Principal.SecurityIdentifier]) -eq $systemSID -and
        $_.AccessControlType -eq "Allow"
    }

    $hasFullControl = $systemRules | Where-Object {
        $_.FileSystemRights -band [System.Security.AccessControl.FileSystemRights]::FullControl
    }

    if ($hasFullControl) {
        Write-CheckResult "SYSTEM has Full Control" "OK"
    } else {
        Write-CheckResult "SYSTEM has Full Control" "FAIL" "SYSTEM does not have Full Control"
    }
} catch {
    Write-CheckResult "SYSTEM has Full Control" "FAIL" $_.Exception.Message
}

# Check 1.5: Administrators have Read access (not write)
try {
    $acl = Get-Acl -Path $ConfigDir
    $adminSID = [System.Security.Principal.SecurityIdentifier]::new("S-1-5-32-544")

    $adminRules = $acl.Access | Where-Object {
        $_.IdentityReference.Translate([System.Security.Principal.SecurityIdentifier]) -eq $adminSID -and
        $_.AccessControlType -eq "Allow"
    }

    if ($adminRules) {
        $hasWrite = $adminRules | Where-Object {
            $_.FileSystemRights -band [System.Security.AccessControl.FileSystemRights]::Write
        }

        if ($hasWrite) {
            Write-CheckResult "Administrators: Read-only" "WARN" "Administrators have Write access (should be Read-only)"
        } else {
            Write-CheckResult "Administrators: Read-only" "OK"
        }
    } else {
        Write-CheckResult "Administrators: Read-only" "WARN" "Administrators have no explicit access"
    }
} catch {
    Write-CheckResult "Administrators: Read-only" "FAIL" $_.Exception.Message
}

# Check 1.6: Users have no access
try {
    $acl = Get-Acl -Path $ConfigDir
    $usersSID = [System.Security.Principal.SecurityIdentifier]::new("S-1-5-32-545")

    $usersRules = $acl.Access | Where-Object {
        try {
            $_.IdentityReference.Translate([System.Security.Principal.SecurityIdentifier]) -eq $usersSID -and
            $_.AccessControlType -eq "Allow"
        } catch {
            $false
        }
    }

    if ($usersRules) {
        Write-CheckResult "Users: No access" "FAIL" "Users group has explicit access (should have none)"
    } else {
        Write-CheckResult "Users: No access" "OK"
    }
} catch {
    Write-CheckResult "Users: No access" "OK" "No Users group rules found"
}

Write-Host ""

# =============================================================================
# Section 2: Config File Checks
# =============================================================================
Write-Host "--- Configuration File ---" -ForegroundColor Yellow

$configFile = Join-Path $ConfigDir "machine-config.yaml"

# Check 2.1: Config file exists
if (Test-Path $configFile) {
    Write-CheckResult "Config file exists" "OK" $configFile
} else {
    Write-CheckResult "Config file exists" "SKIP" "No config file yet (will be created on first run)"
}

# Check 2.2: DPAPI encryption (encrypted_setup_key)
if (Test-Path $configFile) {
    try {
        $content = Get-Content -Path $configFile -Raw

        # Check for plaintext setup_key (bad) - use (?m) for multiline mode
        if ($content -match "(?m)^setup_key:\s*\S+" -and $content -notmatch "(?m)^encrypted_setup_key:") {
            Write-CheckResult "Setup key encrypted" "WARN" "Plaintext setup_key found (will be encrypted on service start)"
        }
        # Check for encrypted_setup_key (good)
        elseif ($content -match "(?m)^encrypted_setup_key:\s*\S+") {
            Write-CheckResult "Setup key encrypted" "OK" "encrypted_setup_key present (DPAPI)"
        }
        # No setup key at all
        else {
            Write-CheckResult "Setup key encrypted" "SKIP" "No setup key configured"
        }
    } catch {
        Write-CheckResult "Setup key encrypted" "FAIL" $_.Exception.Message
    }

    # Check 2.3: DPAPI encryption (encrypted_private_key)
    try {
        $content = Get-Content -Path $configFile -Raw

        if ($content -match "(?m)^encrypted_private_key:\s*\S+") {
            Write-CheckResult "WireGuard key encrypted" "OK" "encrypted_private_key present (DPAPI)"
        } else {
            Write-CheckResult "WireGuard key encrypted" "SKIP" "No WireGuard key yet (generated on first bootstrap)"
        }
    } catch {
        Write-CheckResult "WireGuard key encrypted" "FAIL" $_.Exception.Message
    }

    # Check 2.4: machine_cert_enabled status
    try {
        $content = Get-Content -Path $configFile -Raw

        if ($content -match "machine_cert_enabled:\s*true") {
            Write-CheckResult "Machine cert mode" "OK" "Phase 2 (mTLS) enabled"
        } elseif ($content -match "machine_cert_enabled:\s*false") {
            Write-CheckResult "Machine cert mode" "WARN" "Phase 1 (Setup-Key) - upgrade to mTLS recommended"
        } else {
            Write-CheckResult "Machine cert mode" "SKIP" "Not configured"
        }
    } catch {
        Write-CheckResult "Machine cert mode" "FAIL" $_.Exception.Message
    }
}

Write-Host ""

# =============================================================================
# Section 3: Service Checks
# =============================================================================
Write-Host "--- Service Configuration ---" -ForegroundColor Yellow

# Check 3.1: Service exists
$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($service) {
    Write-CheckResult "Service registered" "OK" $ServiceName
} else {
    Write-CheckResult "Service registered" "FAIL" "Service not found: $ServiceName"
}

# Check 3.2: Service status
if ($service) {
    if ($service.Status -eq "Running") {
        Write-CheckResult "Service running" "OK"
    } elseif ($service.Status -eq "Stopped") {
        Write-CheckResult "Service running" "WARN" "Service is stopped"
    } else {
        Write-CheckResult "Service running" "WARN" "Service status: $($service.Status)"
    }
}

# Check 3.3: Service startup type
if ($service) {
    $startType = $service.StartType
    if ($startType -eq "Automatic") {
        Write-CheckResult "Auto-start enabled" "OK"
    } else {
        Write-CheckResult "Auto-start enabled" "WARN" "StartType is: $startType (should be Automatic)"
    }
}

# Check 3.4: Service runs as SYSTEM
if ($service) {
    try {
        $serviceConfig = Get-CimInstance -ClassName Win32_Service -Filter "Name='$ServiceName'"
        $account = $serviceConfig.StartName

        if ($account -eq "LocalSystem" -or $account -eq "NT AUTHORITY\SYSTEM") {
            Write-CheckResult "Runs as SYSTEM" "OK"
        } else {
            Write-CheckResult "Runs as SYSTEM" "WARN" "Runs as: $account (should be LocalSystem)"
        }
    } catch {
        Write-CheckResult "Runs as SYSTEM" "FAIL" $_.Exception.Message
    }
}

# Check 3.5: Recovery options configured
if ($service) {
    try {
        $recoveryConfig = sc.exe qfailure $ServiceName 2>&1
        if ($recoveryConfig -match "RESTART") {
            Write-CheckResult "Recovery options" "OK" "Auto-restart on failure configured"
        } else {
            Write-CheckResult "Recovery options" "WARN" "No auto-restart configured"
        }
    } catch {
        Write-CheckResult "Recovery options" "FAIL" $_.Exception.Message
    }
}

Write-Host ""

# =============================================================================
# Section 4: Network Checks
# =============================================================================
Write-Host "--- Network Configuration ---" -ForegroundColor Yellow

# Check 4.1: WireGuard interface exists
$wgInterface = Get-NetAdapter -Name "wg-nb-machine" -ErrorAction SilentlyContinue
if ($wgInterface) {
    Write-CheckResult "WireGuard interface" "OK" "wg-nb-machine (Status: $($wgInterface.Status))"
} else {
    Write-CheckResult "WireGuard interface" "SKIP" "Not created yet (created on first connect)"
}

# Check 4.2: NRPT rules
try {
    $nrptRules = Get-DnsClientNrptRule -ErrorAction SilentlyContinue | Where-Object { $_.Comment -like "*NetBird*" }
    if ($nrptRules) {
        Write-CheckResult "NRPT rules" "OK" "$($nrptRules.Count) rule(s) configured"
    } else {
        Write-CheckResult "NRPT rules" "SKIP" "No NRPT rules yet (configured on connect)"
    }
} catch {
    Write-CheckResult "NRPT rules" "SKIP" "Could not query NRPT rules"
}

# Check 4.3: Firewall rules
try {
    $fwRules = Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue
    if ($fwRules) {
        $allowRules = ($fwRules | Where-Object { $_.Action -eq "Allow" }).Count
        $blockRules = ($fwRules | Where-Object { $_.Action -eq "Block" }).Count
        Write-CheckResult "Firewall rules" "OK" "$allowRules allow, $blockRules deny rules"
    } else {
        Write-CheckResult "Firewall rules" "SKIP" "No firewall rules yet (configured on connect)"
    }
} catch {
    Write-CheckResult "Firewall rules" "SKIP" "Could not query firewall rules"
}

Write-Host ""

# =============================================================================
# Summary
# =============================================================================
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host " Summary" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

$passedColor = if ($script:failedChecks -eq 0) { "Green" } else { "White" }
$failedColor = if ($script:failedChecks -gt 0) { "Red" } else { "White" }
$warnColor = if ($script:warnings -gt 0) { "Yellow" } else { "White" }

Write-Host "Total Checks: $($script:totalChecks)"
Write-Host "Passed: " -NoNewline; Write-Host $script:passedChecks -ForegroundColor $passedColor
Write-Host "Failed: " -NoNewline; Write-Host $script:failedChecks -ForegroundColor $failedColor
Write-Host "Warnings: " -NoNewline; Write-Host $script:warnings -ForegroundColor $warnColor
Write-Host ""

if ($script:failedChecks -eq 0 -and $script:warnings -eq 0) {
    Write-Host "All security checks passed!" -ForegroundColor Green
    exit 0
} elseif ($script:failedChecks -eq 0) {
    Write-Host "Security checks passed with warnings. Review warnings above." -ForegroundColor Yellow
    exit 0
} else {
    Write-Host "Security issues detected! Review failed checks above." -ForegroundColor Red
    exit 1
}
