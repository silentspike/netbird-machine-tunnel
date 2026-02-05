# Troubleshooting Guide

This guide covers diagnostics and troubleshooting for the NetBird Machine Tunnel fork features.

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Service Issues](#service-issues)
- [Tunnel Connectivity](#tunnel-connectivity)
- [NRPT (DNS Routing)](#nrpt-dns-routing)
- [Windows Firewall Rules](#windows-firewall-rules)
- [Certificate Issues (mTLS)](#certificate-issues-mtls)
- [Health Monitoring](#health-monitoring)
- [Event Log Reference](#event-log-reference)
- [Log and Config Locations](#log-and-config-locations)

---

## Quick Diagnostics

Run these commands to get a quick overview of the Machine Tunnel status:

```powershell
# Service status
Get-Service NetBirdMachine

# WireGuard interface
Get-NetAdapter | Where-Object { $_.Name -like "wg-nb*" -or $_.Description -like "WireGuard*" }

# NRPT rules (DNS routing)
# Check registry for NetBird NRPT rules (more reliable than Get-DnsClientNrptRule)
Get-ChildItem "HKLM:\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig" -ErrorAction SilentlyContinue |
    Where-Object { $_.PSChildName -like "NetBird-Machine-*" }

# Firewall rules
Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue |
    Format-Table DisplayName, Enabled, Action, Direction

# Recent Event Log entries
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='NetBirdMachine'} -MaxEvents 10 -ErrorAction SilentlyContinue |
    Format-Table TimeCreated, Id, Message -Wrap

# Log file (last 50 lines)
Get-Content "C:\ProgramData\NetBird\machine-tunnel.log" -Tail 50 -ErrorAction SilentlyContinue
```

---

## Service Issues

### Service Does Not Start

**Symptoms:**
- `Get-Service NetBirdMachine` shows `Stopped`
- Service fails immediately after starting

**Diagnostics:**

```powershell
# Check service status details
sc.exe qc NetBirdMachine
sc.exe query NetBirdMachine

# Check Event Log for service errors
Get-WinEvent -FilterHashtable @{LogName='System'; ProviderName='Service Control Manager'} -MaxEvents 20 |
    Where-Object { $_.Message -like "*NetBird*" }

# Verify config file exists and is valid
Test-Path "C:\ProgramData\NetBird\machine-config.yaml"
Get-Content "C:\ProgramData\NetBird\machine-config.yaml"

# Check binary permissions
icacls "C:\Program Files\NetBird Machine\netbird-machine.exe"
```

**Common Causes and Solutions:**

| Cause | Solution |
|-------|----------|
| Config file missing | Create `C:\ProgramData\NetBird\machine-config.yaml` with required settings |
| Invalid YAML syntax | Validate YAML syntax (indentation, quotes) |
| Missing management_url | Add `management_url: "https://your-server:443"` to config |
| Binary not found | Reinstall: `netbird-machine.exe install` |
| Insufficient permissions | Run installation as Administrator |

### Service Starts But Tunnel Does Not Connect

**Diagnostics:**

```powershell
# Check log for connection errors
Select-String -Path "C:\ProgramData\NetBird\machine-tunnel.log" -Pattern "error|failed|timeout" -Context 2

# Verify management server reachability
Test-NetConnection -ComputerName "your-management-server" -Port 443
Test-NetConnection -ComputerName "your-management-server" -Port 33074  # mTLS port

# Check if Setup Key or Certificate is configured
Select-String -Path "C:\ProgramData\NetBird\machine-config.yaml" -Pattern "setup_key|machine_cert"
```

---

## Tunnel Connectivity

### WireGuard Interface Not Created

**Symptoms:**
- No `wg-nb-machine` interface visible
- `Get-NetAdapter` shows no WireGuard adapter

**Diagnostics:**

```powershell
# List all network adapters
Get-NetAdapter | Format-Table Name, InterfaceDescription, Status

# Check for WireGuard driver
Get-WindowsDriver -Online | Where-Object { $_.ProviderName -like "*WireGuard*" }

# Check tunnel log for interface errors
Select-String -Path "C:\ProgramData\NetBird\machine-tunnel.log" -Pattern "interface|wg-nb|wireguard" -Context 2
```

**Common Causes:**
- WireGuard driver not installed
- Another WireGuard instance using the interface
- Insufficient privileges (service must run as SYSTEM)

### Tunnel Connected But DC Not Reachable

**Symptoms:**
- Service shows connected
- `Test-NetConnection -ComputerName dc01.domain.local -Port 389` fails

**Diagnostics:**

```powershell
# Check route to DC network
route print | Select-String "192.168"

# Verify DNS resolution works
Resolve-DnsName dc01.domain.local

# Test connectivity step by step
Test-NetConnection -ComputerName 192.168.100.20 -Port 53   # DNS
Test-NetConnection -ComputerName 192.168.100.20 -Port 88   # Kerberos
Test-NetConnection -ComputerName 192.168.100.20 -Port 389  # LDAP
Test-NetConnection -ComputerName 192.168.100.20 -Port 445  # SMB

# Check if traffic goes through tunnel interface
Get-NetRoute | Where-Object { $_.DestinationPrefix -like "192.168.100.*" }
```

**Common Causes:**

| Cause | Solution |
|-------|----------|
| Route missing | Check NetBird Dashboard: Route for DC subnet configured? |
| NRPT not active | Restart service, verify NRPT rules (see below) |
| Firewall blocking | Check Windows Firewall rules (see below) |
| Router-Peer down | Verify Router-Peer VM is running and NetBird connected |
| Return route missing | On DC network: Add route to 100.64.0.0/10 via Router-Peer |

---

## NRPT (DNS Routing)

The Machine Tunnel uses Windows Name Resolution Policy Table (NRPT) to route AD domain DNS queries through the tunnel.

### Check NRPT Rules

```powershell
# List all NRPT rules
Get-DnsClientNrptRule

# List only NetBird Machine Tunnel rules
# Check registry for NetBird NRPT rules (more reliable than Get-DnsClientNrptRule)
Get-ChildItem "HKLM:\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig" -ErrorAction SilentlyContinue |
    Where-Object { $_.PSChildName -like "NetBird-Machine-*" }

# Check registry directly (authoritative)
Get-ChildItem "HKLM:\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig" |
    Where-Object { $_.PSChildName -like "NetBird-Machine-*" }

# View rule details
reg query "HKLM\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig" /s | findstr /i "netbird"
```

### NRPT Rules Missing

**Symptoms:**
- `Get-DnsClientNrptRule` shows no NetBird rules
- DNS queries for AD domain go to wrong DNS server

**Diagnostics:**

```powershell
# Check if DNS Client service is running
Get-Service Dnscache

# Check log for NRPT errors
Select-String -Path "C:\ProgramData\NetBird\machine-tunnel.log" -Pattern "NRPT|nrpt|dns" -Context 2

# Verify config has NRPT settings
Select-String -Path "C:\ProgramData\NetBird\machine-config.yaml" -Pattern "nrpt"
```

**Solutions:**

1. Restart the service: `Restart-Service NetBirdMachine`
2. Restart DNS Client: `Restart-Service Dnscache`
3. Verify config contains `nrpt_domains` and `nrpt_dns_servers`

### DNS Resolution Uses Wrong Server

**Diagnostics:**

```powershell
# Check which DNS server is used
Resolve-DnsName dc01.corp.local -DnsOnly | Format-List

# Clear DNS cache and retry
Clear-DnsClientCache
ipconfig /flushdns
Resolve-DnsName dc01.corp.local

# Check NRPT namespace matches
Get-DnsClientNrptRule | Select-Object Namespace, NameServers
# Namespace should start with "." (e.g., ".corp.local")
```

### Manually Remove NRPT Rules (Cleanup)

```powershell
# Remove only NetBird Machine Tunnel rules (safe)
Get-ChildItem "HKLM:\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig" |
    Where-Object { $_.PSChildName -like "NetBird-Machine-*" } |
    ForEach-Object { Remove-Item $_.PSPath -Recurse }

# Restart DNS Client to apply
Restart-Service Dnscache
```

---

## Windows Firewall Rules

The Machine Tunnel creates Windows Firewall rules to allow DC traffic on the tunnel interface while blocking other traffic (defense in depth).

### Check Firewall Rules

```powershell
# List all NetBird Machine Tunnel rules
Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue |
    Format-Table DisplayName, Enabled, Action, Direction

# Get detailed rule info including ports
Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue |
    Get-NetFirewallPortFilter |
    Format-Table @{L='Rule';E={$_.InstanceID}}, Protocol, LocalPort, RemotePort

# Get IP address filters
Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue |
    Get-NetFirewallAddressFilter |
    Format-Table @{L='Rule';E={$_.InstanceID}}, LocalAddress, RemoteAddress

# Alternative: Use netsh
netsh advfirewall firewall show rule group="NetBird Machine Tunnel"
```

### Firewall Rules Missing

**Symptoms:**
- No rules in "NetBird Machine Tunnel" group
- DC traffic blocked even with tunnel connected

**Diagnostics:**

```powershell
# Check log for firewall errors
Select-String -Path "C:\ProgramData\NetBird\machine-tunnel.log" -Pattern "firewall|rule" -Context 2

# Verify config has firewall settings
Select-String -Path "C:\ProgramData\NetBird\machine-config.yaml" -Pattern "firewall"
```

**Solutions:**

1. Restart the service: `Restart-Service NetBirdMachine`
2. Verify config contains `firewall_dc_ips` with DC IP addresses

### Manually Remove Firewall Rules (Cleanup)

```powershell
# Remove all NetBird Machine Tunnel rules (safe)
Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue |
    Remove-NetFirewallRule

# Verify removal
Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue
```

---

## Certificate Issues (mTLS)

Machine certificates are used for mTLS authentication after the initial bootstrap phase.

### Check Machine Certificate

```powershell
# List machine certificates
Get-ChildItem Cert:\LocalMachine\My | Format-Table Subject, Thumbprint, NotAfter

# Find certificate matching this machine
$cert = Get-ChildItem Cert:\LocalMachine\My |
    Where-Object { $_.Subject -match $env:COMPUTERNAME }
$cert | Format-List *

# Check certificate SAN (Subject Alternative Name) - CRITICAL for mTLS!
$cert.Extensions | Where-Object { $_.Oid.FriendlyName -eq "Subject Alternative Name" } |
    ForEach-Object { $_.Format($true) }
# Must contain: DNS Name=hostname.domain.local

# Check certificate validity
$cert | Select-Object Subject, NotBefore, NotAfter,
    @{N='DaysRemaining';E={($_.NotAfter - (Get-Date)).Days}}

# Verify certificate chain
certutil -verify -urlfetch $cert.Thumbprint
```

### mTLS Authentication Fails

**Symptoms:**
- Log shows "mTLS validation failed" or "certificate rejected"
- Connection to port 33074 fails

**Diagnostics:**

```powershell
# Check if certificate has correct SAN
$cert = Get-ChildItem Cert:\LocalMachine\My | Where-Object { $_.Subject -match $env:COMPUTERNAME }
$san = $cert.Extensions | Where-Object { $_.Oid.Value -eq "2.5.29.17" }
$san.Format($true)
# Expected: DNS Name=HOSTNAME.domain.local

# Check if certificate is from correct template
$cert.Extensions | Where-Object { $_.Oid.FriendlyName -eq "Certificate Template Information" } |
    ForEach-Object { $_.Format($true) }

# Test mTLS connection (requires OpenSSL or similar)
# From management server, check logs for rejection reason
```

**Common Causes:**

| Cause | Solution |
|-------|----------|
| SAN missing DNS name | Reconfigure AD CS template to include DNS name in SAN |
| Wrong domain in SAN | SAN must match `hostname.domain` exactly |
| Certificate expired | Renew certificate via AD CS |
| CA not trusted by server | Add CA certificate to server's `mtls_ca_file` |
| Domain not in AllowedDomains | Add domain to server's per-account AllowedDomains |

### Certificate Enrollment

```powershell
# Request new machine certificate (if Auto-Enrollment configured)
certreq -autoenroll

# Manual enrollment with specific template
certreq -enroll -machine "NetBirdMachine"

# Check enrollment status
certutil -pulse
```

---

## Health Monitoring

The Machine Tunnel includes health monitoring with automatic reconnection.

### Check Health Status

```powershell
# Check log for health status
Select-String -Path "C:\ProgramData\NetBird\machine-tunnel.log" -Pattern "health|healthy|degraded|unhealthy" -Context 1

# Check for reconnection events
Select-String -Path "C:\ProgramData\NetBird\machine-tunnel.log" -Pattern "reconnect|backoff" -Context 2
```

### Health Check Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| Check Interval | 30 seconds | Time between health checks |
| Handshake Timeout | 2 minutes | Max age of last WireGuard handshake |
| gRPC Ping Timeout | 10 seconds | Timeout for management server ping |
| Consecutive Failures | 3 | Failures before triggering reconnect |

### Health States

| State | Description |
|-------|-------------|
| `healthy` | All checks pass |
| `degraded` | Some checks fail but tunnel functional |
| `unhealthy` | Tunnel not functional, reconnect triggered |
| `unknown` | Initial state or check not completed |

---

## Event Log Reference

The Machine Tunnel logs events to Windows Event Log under source `NetBirdMachine`.

### View Events

```powershell
# View all NetBirdMachine events
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='NetBirdMachine'} -MaxEvents 50 |
    Format-Table TimeCreated, Id, LevelDisplayName, Message -Wrap

# Filter by Event ID
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='NetBirdMachine'; Id=1100} -MaxEvents 10

# Filter by severity (Error=2, Warning=3, Info=4)
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='NetBirdMachine'; Level=2} -MaxEvents 10
```

### Event IDs

| Event ID | Description | Severity |
|----------|-------------|----------|
| 1000 | Service Start | Info |
| 1001 | Service Stop | Info |
| 1100 | Tunnel Connected | Info |
| 1101 | Tunnel Disconnected | Warning |
| 1200 | Authentication Success | Info |
| 1201 | Authentication Failure | Error |
| 1300 | ACL Hardened (Firewall rules applied) | Info |
| 1301 | Setup Key Removed (mTLS transition complete) | Info |
| 1400 | Configuration Error | Error |

### Register Event Source (if missing)

```powershell
# Run as Administrator
New-EventLog -LogName Application -Source "NetBirdMachine" -ErrorAction SilentlyContinue
```

---

## Log and Config Locations

### File Locations

| File | Path | Description |
|------|------|-------------|
| Config | `C:\ProgramData\NetBird\machine-config.yaml` | Configuration with DPAPI-encrypted secrets |
| Log File | `C:\ProgramData\NetBird\machine-tunnel.log` | Application log |
| Binary | (install location) | Service executable (path used during `install`) |

### Registry Locations

| Key | Path | Description |
|-----|------|-------------|
| NRPT Rules | `HKLM:\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig` | DNS policy rules |
| Service Config | `HKLM:\SYSTEM\CurrentControlSet\Services\NetBirdMachine` | Service registration |

### Log Levels

Set in `machine-config.yaml`:

```yaml
log_level: "debug"  # debug | info | warn | error
```

### Increase Log Verbosity for Debugging

```powershell
# Edit config
notepad C:\ProgramData\NetBird\machine-config.yaml
# Change: log_level: "debug"

# Restart service
Restart-Service NetBirdMachine

# Watch log in real-time
Get-Content "C:\ProgramData\NetBird\machine-tunnel.log" -Wait -Tail 50
```

---

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture and design decisions
- [README.md](../README.md) - Quick start guide
- [NOTICE.md](../NOTICE.md) - License and attribution
