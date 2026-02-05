# Troubleshooting Guide

This guide covers diagnostics for the NetBird Machine Tunnel. All commands have been tested on Windows 11.

## Quick Status Check

```powershell
# Service status
Get-Service NetBirdMachine

# WireGuard interface
Get-NetAdapter | Where-Object { $_.Name -like "wg*" }

# DC connectivity (replace IP with your DC)
Test-NetConnection -ComputerName 192.168.100.20 -Port 389 -WarningAction SilentlyContinue
```

**Expected output when healthy:**
```
Name            : NetBirdMachine
Status          : Running

Name          Status InterfaceDescription
----          ------ --------------------
wg-nb-machine Up     WireGuard Tunnel

ComputerName     : 192.168.100.20
TcpTestSucceeded : True
```

---

## Service Issues

### Check Service Status

```powershell
Get-Service NetBirdMachine | Format-List Name, Status, StartType
```

### Start/Stop Service

```powershell
# Start
Start-Service NetBirdMachine

# Stop
Stop-Service NetBirdMachine

# Restart
Restart-Service NetBirdMachine
```

### View Service Events

```powershell
Get-WinEvent -FilterHashtable @{LogName="Application"; ProviderName="NetBirdMachine"} -MaxEvents 10 -ErrorAction SilentlyContinue |
    Format-Table TimeCreated, Id, Message -Wrap
```

---

## Tunnel Connectivity

### Check WireGuard Interface

```powershell
# Interface status
Get-NetAdapter | Where-Object { $_.Name -like "wg*" -or $_.InterfaceDescription -like "*WireGuard*" } |
    Format-Table Name, Status, InterfaceDescription

# Interface IP address
Get-NetIPAddress -InterfaceAlias "wg-nb-machine" -ErrorAction SilentlyContinue
```

### Test DC Connectivity

Test connectivity to Domain Controller ports:

```powershell
# DNS (53)
Test-NetConnection -ComputerName 192.168.100.20 -Port 53 -WarningAction SilentlyContinue

# Kerberos (88)
Test-NetConnection -ComputerName 192.168.100.20 -Port 88 -WarningAction SilentlyContinue

# LDAP (389)
Test-NetConnection -ComputerName 192.168.100.20 -Port 389 -WarningAction SilentlyContinue

# SMB (445)
Test-NetConnection -ComputerName 192.168.100.20 -Port 445 -WarningAction SilentlyContinue
```

Replace `192.168.100.20` with your DC IP address.

---

## Log Files

### File Locations

| File | Path |
|------|------|
| Config | `C:\ProgramData\NetBird\machine-config.yaml` |
| Log | `C:\ProgramData\NetBird\machine-tunnel.log` |

### Check Files Exist

```powershell
Test-Path "C:\ProgramData\NetBird\machine-config.yaml"
Test-Path "C:\ProgramData\NetBird\machine-tunnel.log"
```

### List All NetBird Files

```powershell
Get-ChildItem "C:\ProgramData\NetBird\" -ErrorAction SilentlyContinue
```

### View Recent Log Entries

```powershell
# Last 50 lines
Get-Content "C:\ProgramData\NetBird\machine-tunnel.log" -Tail 50

# Follow log in real-time
Get-Content "C:\ProgramData\NetBird\machine-tunnel.log" -Wait -Tail 20
```

### Search Log for Errors

```powershell
Select-String -Path "C:\ProgramData\NetBird\machine-tunnel.log" -Pattern "error|failed|timeout" | Select-Object -Last 20
```

---

## DNS (NRPT)

### Check NRPT Rules

```powershell
# List all NRPT rules
Get-DnsClientNrptRule

# Check registry directly
Get-ChildItem "HKLM:\SYSTEM\CurrentControlSet\Services\Dnscache\Parameters\DnsPolicyConfig" -ErrorAction SilentlyContinue
```

### Test DNS Resolution

```powershell
# Resolve DC hostname
Resolve-DnsName dc01.test.local -ErrorAction SilentlyContinue

# Check which DNS server is used
Resolve-DnsName dc01.test.local -DnsOnly -ErrorAction SilentlyContinue | Format-List
```

### Clear DNS Cache

```powershell
Clear-DnsClientCache
ipconfig /flushdns
```

---

## Firewall

### Check NetBird Firewall Rules

```powershell
Get-NetFirewallRule -DisplayName "*NetBird*" -ErrorAction SilentlyContinue |
    Format-Table DisplayName, Enabled, Direction, Action
```

---

## Common Issues

### Service Won't Start

1. Check config file exists:
   ```powershell
   Test-Path "C:\ProgramData\NetBird\machine-config.yaml"
   ```

2. Check Event Log for errors:
   ```powershell
   Get-WinEvent -FilterHashtable @{LogName="Application"; ProviderName="NetBirdMachine"} -MaxEvents 5
   ```

3. Check log file for startup errors:
   ```powershell
   Get-Content "C:\ProgramData\NetBird\machine-tunnel.log" -Tail 100 | Select-String "error|failed"
   ```

### Tunnel Connected But DC Not Reachable

1. Verify interface is up:
   ```powershell
   Get-NetAdapter -Name "wg-nb-machine"
   ```

2. Check route exists:
   ```powershell
   Get-NetRoute | Where-Object { $_.DestinationPrefix -like "192.168.100.*" }
   ```

3. Test connectivity step by step:
   ```powershell
   Test-NetConnection -ComputerName 192.168.100.20 -Port 389 -WarningAction SilentlyContinue
   ```

### DNS Not Resolving AD Names

1. Check NRPT rules exist:
   ```powershell
   Get-DnsClientNrptRule
   ```

2. Clear DNS cache:
   ```powershell
   Clear-DnsClientCache
   ```

3. Test resolution:
   ```powershell
   Resolve-DnsName dc01.test.local
   ```

---

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [NOTICE.md](../NOTICE.md) - License and attribution
