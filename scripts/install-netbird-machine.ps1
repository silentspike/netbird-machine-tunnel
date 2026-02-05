#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Installs the NetBird Machine Tunnel Service with hardened permissions.

.DESCRIPTION
    This script performs a secure installation of the NetBird Machine Tunnel Service:
    - Creates config directory with hardened ACLs (SYSTEM=Full, Admins=Read, Users=None)
    - Copies binary to installation directory with proper permissions
    - Registers the Windows Service with SCM
    - Configures service recovery options

    Enterprise SOTA: Uses Windows Security Descriptors and explicit ACL management
    to ensure configuration files are protected from unauthorized access.

.PARAMETER BinaryPath
    Path to the netbird-machine.exe binary to install.

.PARAMETER InstallDir
    Installation directory for the binary. Default: C:\Program Files\NetBird Machine

.PARAMETER ConfigDir
    Configuration directory. Default: C:\ProgramData\NetBird

.PARAMETER ServiceName
    Name of the Windows Service. Default: NetBirdMachine

.PARAMETER ManagementURL
    URL of the NetBird management server (required for initial config).

.PARAMETER SetupKey
    Setup key for initial bootstrap (optional, can be added later).

.PARAMETER WhatIf
    Shows what would be done without making changes.

.EXAMPLE
    .\install-netbird-machine.ps1 -BinaryPath .\netbird-machine.exe -ManagementURL https://netbird.example.com

.EXAMPLE
    .\install-netbird-machine.ps1 -BinaryPath .\netbird-machine.exe -ManagementURL https://netbird.example.com -SetupKey "abc123" -WhatIf

.NOTES
    Author: NetBird Machine Tunnel Fork
    Requires: Windows 10/11, Server 2019+, Administrator privileges
    Version: 1.0.0
#>

[CmdletBinding(SupportsShouldProcess)]
param(
    [Parameter(Mandatory = $true)]
    [ValidateScript({ Test-Path $_ -PathType Leaf })]
    [string]$BinaryPath,

    [Parameter(Mandatory = $false)]
    [string]$InstallDir = "C:\Program Files\NetBird Machine",

    [Parameter(Mandatory = $false)]
    [string]$ConfigDir = "C:\ProgramData\NetBird",

    [Parameter(Mandatory = $false)]
    [string]$ServiceName = "NetBirdMachine",

    [Parameter(Mandatory = $true)]
    [ValidatePattern('^https?://')]
    [string]$ManagementURL,

    [Parameter(Mandatory = $false)]
    [string]$SetupKey = ""
)

# Strict mode for better error handling
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Logging helper
function Write-Log {
    param([string]$Message, [string]$Level = "INFO")
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    $color = switch ($Level) {
        "INFO"    { "White" }
        "SUCCESS" { "Green" }
        "WARNING" { "Yellow" }
        "ERROR"   { "Red" }
        default   { "White" }
    }
    Write-Host "[$timestamp] [$Level] $Message" -ForegroundColor $color
}

# Check if running as Administrator
function Test-Administrator {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($identity)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

if (-not (Test-Administrator)) {
    Write-Log "This script requires Administrator privileges. Please run as Administrator." "ERROR"
    exit 1
}

Write-Log "=== NetBird Machine Tunnel Installation ===" "INFO"
Write-Log "Binary: $BinaryPath" "INFO"
Write-Log "Install Directory: $InstallDir" "INFO"
Write-Log "Config Directory: $ConfigDir" "INFO"
Write-Log "Service Name: $ServiceName" "INFO"
Write-Log "Management URL: $ManagementURL" "INFO"

# Step 1: Create installation directory
Write-Log "Step 1: Creating installation directory..." "INFO"
if ($PSCmdlet.ShouldProcess($InstallDir, "Create directory")) {
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Write-Log "Created: $InstallDir" "SUCCESS"
    } else {
        Write-Log "Already exists: $InstallDir" "INFO"
    }
}

# Step 2: Create config directory with hardened ACLs
Write-Log "Step 2: Creating config directory with hardened ACLs..." "INFO"
if ($PSCmdlet.ShouldProcess($ConfigDir, "Create directory with hardened ACLs")) {
    if (-not (Test-Path $ConfigDir)) {
        New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
        Write-Log "Created: $ConfigDir" "SUCCESS"
    }

    # Get SIDs for SYSTEM and Administrators
    $systemSID = [System.Security.Principal.SecurityIdentifier]::new("S-1-5-18")
    $adminSID = [System.Security.Principal.SecurityIdentifier]::new("S-1-5-32-544")

    # Create new ACL
    $acl = New-Object System.Security.AccessControl.DirectorySecurity

    # Disable inheritance and remove inherited rules
    $acl.SetAccessRuleProtection($true, $false)

    # Add SYSTEM: Full Control (includes inheritance to child objects)
    $systemRule = New-Object System.Security.AccessControl.FileSystemAccessRule(
        $systemSID,
        [System.Security.AccessControl.FileSystemRights]::FullControl,
        ([System.Security.AccessControl.InheritanceFlags]::ContainerInherit -bor [System.Security.AccessControl.InheritanceFlags]::ObjectInherit),
        [System.Security.AccessControl.PropagationFlags]::None,
        [System.Security.AccessControl.AccessControlType]::Allow
    )
    $acl.AddAccessRule($systemRule)

    # Add Administrators: Read Only (for troubleshooting)
    $adminRule = New-Object System.Security.AccessControl.FileSystemAccessRule(
        $adminSID,
        [System.Security.AccessControl.FileSystemRights]::ReadAndExecute,
        ([System.Security.AccessControl.InheritanceFlags]::ContainerInherit -bor [System.Security.AccessControl.InheritanceFlags]::ObjectInherit),
        [System.Security.AccessControl.PropagationFlags]::None,
        [System.Security.AccessControl.AccessControlType]::Allow
    )
    $acl.AddAccessRule($adminRule)

    # Set owner to SYSTEM
    $acl.SetOwner($systemSID)

    # Apply ACL
    Set-Acl -Path $ConfigDir -AclObject $acl
    Write-Log "Applied hardened ACLs to: $ConfigDir" "SUCCESS"
    Write-Log "  - SYSTEM: Full Control" "INFO"
    Write-Log "  - Administrators: Read & Execute" "INFO"
    Write-Log "  - Users: No Access" "INFO"
}

# Step 3: Copy binary to installation directory
Write-Log "Step 3: Copying binary to installation directory..." "INFO"
$targetBinary = Join-Path $InstallDir "netbird-machine.exe"
if ($PSCmdlet.ShouldProcess($targetBinary, "Copy binary")) {
    Copy-Item -Path $BinaryPath -Destination $targetBinary -Force
    Write-Log "Copied binary to: $targetBinary" "SUCCESS"
}

# Step 4: Create initial config file (if setup key provided)
$configFile = Join-Path $ConfigDir "machine-config.yaml"
if ($SetupKey -and $PSCmdlet.ShouldProcess($configFile, "Create initial config")) {
    Write-Log "Step 4: Creating initial configuration..." "INFO"

    # Note: The actual DPAPI encryption happens when the service starts.
    # This creates a plaintext setup_key which will be encrypted on first run.
    $configContent = @"
# NetBird Machine Tunnel Configuration
# Generated by install-netbird-machine.ps1
# WARNING: setup_key will be encrypted by the service on first run

management_url: $ManagementURL
setup_key: $SetupKey
machine_cert_enabled: false
"@

    Set-Content -Path $configFile -Value $configContent -Force
    Write-Log "Created initial config: $configFile" "SUCCESS"
    Write-Log "  Note: setup_key will be DPAPI-encrypted on first service start" "WARNING"
} elseif (-not $SetupKey) {
    Write-Log "Step 4: Skipping config creation (no SetupKey provided)" "INFO"
}

# Step 5: Stop existing service if running
Write-Log "Step 5: Checking for existing service..." "INFO"
$existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($existingService) {
    if ($existingService.Status -eq "Running") {
        if ($PSCmdlet.ShouldProcess($ServiceName, "Stop service")) {
            Write-Log "Stopping existing service..." "INFO"
            Stop-Service -Name $ServiceName -Force
            Start-Sleep -Seconds 2
        }
    }
    if ($PSCmdlet.ShouldProcess($ServiceName, "Remove existing service")) {
        Write-Log "Removing existing service registration..." "INFO"
        sc.exe delete $ServiceName | Out-Null
        Start-Sleep -Seconds 1
    }
}

# Step 6: Register Windows Service
Write-Log "Step 6: Registering Windows Service..." "INFO"
$serviceBinary = "`"$targetBinary`" run --config `"$configFile`" --log-level info --log-file `"$ConfigDir\machine-tunnel.log`""

if ($PSCmdlet.ShouldProcess($ServiceName, "Create Windows Service")) {
    # Create service
    $result = sc.exe create $ServiceName `
        binPath= $serviceBinary `
        start= auto `
        DisplayName= "NetBird Machine Tunnel" `
        obj= "LocalSystem"

    if ($LASTEXITCODE -ne 0) {
        Write-Log "Failed to create service: $result" "ERROR"
        exit 1
    }
    Write-Log "Service registered: $ServiceName" "SUCCESS"

    # Set service description
    sc.exe description $ServiceName "NetBird Machine Tunnel - Windows Pre-Login VPN for Enterprise" | Out-Null

    # Configure recovery options (restart on failure)
    # reset= 86400 = reset failure count after 1 day
    # actions= restart/5000/restart/10000/restart/30000 = restart after 5s, 10s, 30s
    sc.exe failure $ServiceName reset= 86400 actions= restart/5000/restart/10000/restart/30000 | Out-Null
    Write-Log "Configured service recovery options (auto-restart on failure)" "SUCCESS"
}

# Step 7: Start service
Write-Log "Step 7: Starting service..." "INFO"
if ($PSCmdlet.ShouldProcess($ServiceName, "Start service")) {
    Start-Service -Name $ServiceName
    Start-Sleep -Seconds 3

    $service = Get-Service -Name $ServiceName
    if ($service.Status -eq "Running") {
        Write-Log "Service started successfully!" "SUCCESS"
    } else {
        Write-Log "Service failed to start. Check logs at: $ConfigDir\machine-tunnel.log" "ERROR"
        exit 1
    }
}

# Summary
Write-Log "" "INFO"
Write-Log "=== Installation Complete ===" "SUCCESS"
Write-Log "Binary: $targetBinary" "INFO"
Write-Log "Config: $configFile" "INFO"
Write-Log "Logs: $ConfigDir\machine-tunnel.log" "INFO"
Write-Log "Service: $ServiceName (Status: $((Get-Service -Name $ServiceName).Status))" "INFO"
Write-Log "" "INFO"

if ($SetupKey) {
    Write-Log "IMPORTANT: After Domain-Join and Certificate Enrollment:" "WARNING"
    Write-Log "  1. Update config to set machine_cert_enabled: true" "WARNING"
    Write-Log "  2. REVOKE the setup-key in NetBird Dashboard!" "WARNING"
}

Write-Log "" "INFO"
Write-Log "Verify installation with: .\verify-config-hardening.ps1" "INFO"
