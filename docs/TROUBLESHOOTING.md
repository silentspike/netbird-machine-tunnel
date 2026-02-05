# Troubleshooting Guide

This guide helps you diagnose and fix problems with the NetBird Machine Tunnel. It covers both the Windows client and the management server.

All commands in this guide have been tested in our lab environment.

---

## Table of Contents

1. [Understanding the System](#understanding-the-system)
2. [Quick Health Check](#quick-health-check)
3. [Client Troubleshooting (Windows)](#client-troubleshooting-windows)
4. [Server Troubleshooting (Management)](#server-troubleshooting-management)
5. [Common Problems and Solutions](#common-problems-and-solutions)

---

## Understanding the System

Before troubleshooting, it helps to understand how the Machine Tunnel works:

```
┌─────────────────┐         ┌─────────────────┐         ┌─────────────────┐
│  Windows Client │ ──────► │   Management    │ ──────► │ Domain          │
│  (Your PC)      │  mTLS   │   Server        │         │ Controller      │
│                 │         │   Port 33074    │         │                 │
│  Service:       │         │                 │         │  DNS, Kerberos  │
│  NetBirdMachine │         │  Validates      │         │  LDAP, etc.     │
│                 │         │  certificates   │         │                 │
└─────────────────┘         └─────────────────┘         └─────────────────┘
        │                                                       ▲
        │                                                       │
        └───────────────── WireGuard Tunnel ───────────────────┘
```

**Key components:**

| Component | What it does | Where to check |
|-----------|--------------|----------------|
| **NetBirdMachine Service** | Runs the VPN tunnel on Windows | Windows Services |
| **WireGuard Interface** | Creates the encrypted tunnel | Network Adapters |
| **Management Server** | Authenticates clients, assigns IPs | Docker logs |
| **mTLS (Port 33074)** | Secure authentication with certificates | Server logs |

---

## Quick Health Check

Run these checks to get an overview of your system's health.

### On the Windows Client

Open PowerShell as Administrator and run:

```powershell
# 1. Is the service running?
Get-Service NetBirdMachine

# 2. Is the tunnel interface active?
Get-NetAdapter | Where-Object { $_.Name -like "wg*" }

# 3. Can you reach the Domain Controller?
Test-NetConnection -ComputerName 192.168.100.20 -Port 389 -WarningAction SilentlyContinue
```

**What healthy output looks like:**

```
Status   Name
------   ----
Running  NetBirdMachine

Name          Status
----          ------
wg-nb-machine Up

TcpTestSucceeded : True
```

### On the Management Server

Connect to your management server and run:

```bash
# Check if management container is running
docker ps | grep management

# Check recent logs for errors
docker logs ubuntu-management-1 --tail 50 | grep -E "ERROR|WARN|mTLS"
```

**What healthy output looks like:**

```
INFO server/mtls_server.go:106: starting mTLS-only Machine Tunnel server on port 33074
INFO grpc/machine_tunnel.go:202: SyncMachinePeer: DNS=your-computer.domain.local
```

---

## Client Troubleshooting (Windows)

### Step 1: Check the Service

The NetBirdMachine service must be running for the tunnel to work.

```powershell
# Check service status
Get-Service NetBirdMachine | Format-List Name, Status, StartType
```

**If the service is stopped:**

```powershell
# Try to start it
Start-Service NetBirdMachine

# Check if it started successfully
Get-Service NetBirdMachine
```

**If the service won't start**, check the Windows Event Log:

```powershell
Get-WinEvent -FilterHashtable @{LogName="Application"; ProviderName="NetBirdMachine"} -MaxEvents 10 -ErrorAction SilentlyContinue |
    Format-Table TimeCreated, Id, Message -Wrap
```

Common reasons the service won't start:
- Configuration file is missing or invalid
- Another VPN software is conflicting
- Windows Firewall is blocking required ports

### Step 2: Check the Tunnel Interface

The service creates a WireGuard network interface called `wg-nb-machine`.

```powershell
# List all network adapters - look for the WireGuard tunnel
Get-NetAdapter | Where-Object { $_.Name -like "wg*" -or $_.InterfaceDescription -like "*WireGuard*" }
```

**If the interface is missing:**
- The service might not be running
- There could be a driver problem
- Check the log file for errors

**If the interface shows "Disconnected":**
- The tunnel couldn't connect to the management server
- Check your internet connection
- Verify the management server is reachable

### Step 3: Check the Log File

The log file contains detailed information about what the service is doing.

**Where is the log file?**

```
C:\ProgramData\NetBird\machine-tunnel.log
```

**View the most recent entries:**

```powershell
Get-Content "C:\ProgramData\NetBird\machine-tunnel.log" -Tail 50
```

**Search for errors:**

```powershell
Select-String -Path "C:\ProgramData\NetBird\machine-tunnel.log" -Pattern "error|failed|timeout" |
    Select-Object -Last 20
```

**Watch the log in real-time** (useful when reproducing a problem):

```powershell
Get-Content "C:\ProgramData\NetBird\machine-tunnel.log" -Wait -Tail 20
```

### Step 4: Check Network Connectivity

Test if you can reach the Domain Controller through the tunnel.

```powershell
# Test DNS (port 53)
Test-NetConnection -ComputerName 192.168.100.20 -Port 53 -WarningAction SilentlyContinue

# Test Kerberos (port 88)
Test-NetConnection -ComputerName 192.168.100.20 -Port 88 -WarningAction SilentlyContinue

# Test LDAP (port 389)
Test-NetConnection -ComputerName 192.168.100.20 -Port 389 -WarningAction SilentlyContinue

# Test SMB for Group Policy (port 445)
Test-NetConnection -ComputerName 192.168.100.20 -Port 445 -WarningAction SilentlyContinue
```

Replace `192.168.100.20` with your Domain Controller's IP address.

**If connectivity fails:**
- Check if the tunnel interface is up
- Verify the route to the DC network exists
- Make sure the Router-Peer is running and forwarding traffic

### Step 5: Check DNS Resolution

The Machine Tunnel uses NRPT (Name Resolution Policy Table) to route DNS queries for your AD domain through the tunnel.

```powershell
# List all NRPT rules
Get-DnsClientNrptRule

# Test if AD DNS names resolve correctly
Resolve-DnsName dc01.yourdomain.local
```

**If DNS doesn't work:**

```powershell
# Clear the DNS cache and try again
Clear-DnsClientCache
ipconfig /flushdns
Resolve-DnsName dc01.yourdomain.local
```

### Step 6: Check the Configuration File

The configuration is stored in:

```
C:\ProgramData\NetBird\machine-config.yaml
```

**Verify the file exists:**

```powershell
Test-Path "C:\ProgramData\NetBird\machine-config.yaml"
```

**List all NetBird files:**

```powershell
Get-ChildItem "C:\ProgramData\NetBird\" -ErrorAction SilentlyContinue
```

---

## Server Troubleshooting (Management)

The management server runs in Docker and handles client authentication.

### Step 1: Check if the Server is Running

```bash
# List running containers
docker ps | grep management

# You should see something like:
# ubuntu-management-1   Up 2 days
```

**If the container is not running:**

```bash
# Check container status
docker ps -a | grep management

# View container logs to see why it stopped
docker logs ubuntu-management-1
```

### Step 2: Check Server Logs

The server logs show authentication attempts and errors.

```bash
# View recent logs
docker logs ubuntu-management-1 --tail 100

# Filter for mTLS-related messages
docker logs ubuntu-management-1 --tail 100 | grep -i "mtls"

# Filter for errors and warnings
docker logs ubuntu-management-1 --tail 100 | grep -E "ERROR|WARN"
```

**What to look for in the logs:**

| Log Message | What it means |
|-------------|---------------|
| `mTLS server configured: port=33074` | Server started correctly |
| `SyncMachinePeer: DNS=...` | A client connected successfully |
| `no peer auth method provided` | Client tried to connect without proper authentication |
| `certificate validation failed` | Client certificate was rejected |

### Step 3: Check mTLS Configuration

The server needs a CA certificate to validate client certificates.

```bash
# Check if the mTLS server is listening (run on the host, not in container)
ss -tlnp | grep 33074
```

**Expected output:**

```
LISTEN  0  4096  0.0.0.0:33074  0.0.0.0:*
```

### Step 4: Test Client Connection

You can check if a specific client is connecting by searching the logs:

```bash
# Search for a specific computer name
docker logs ubuntu-management-1 --tail 500 | grep "your-computer-name"
```

---

## Common Problems and Solutions

### Problem: Service starts but tunnel doesn't connect

**Symptoms:**
- Service shows as "Running"
- But no `wg-nb-machine` interface appears
- Or interface appears but stays "Disconnected"

**Possible causes:**

1. **Can't reach management server**
   - Check if port 443 and 33074 are open in your firewall
   - Test: `Test-NetConnection -ComputerName your-management-server -Port 33074`

2. **Certificate problem**
   - The machine certificate might be missing or invalid
   - Check: `Get-ChildItem Cert:\LocalMachine\My`

3. **Configuration problem**
   - Check the log file for specific errors
   - Verify `machine-config.yaml` exists and is valid

### Problem: Tunnel connects but can't reach Domain Controller

**Symptoms:**
- Service running, interface up
- But `Test-NetConnection` to DC fails

**Possible causes:**

1. **Router-Peer not running**
   - The Router-Peer VM forwards traffic to the DC network
   - Check if it's running and NetBird is connected

2. **Route missing**
   - Check: `Get-NetRoute | Where-Object { $_.DestinationPrefix -like "192.168.*" }`
   - The DC network should route through the WireGuard interface

3. **Firewall blocking traffic**
   - Check Windows Firewall rules
   - Check: `Get-NetFirewallRule -DisplayName "*NetBird*"`

### Problem: DNS not resolving AD domain names

**Symptoms:**
- Tunnel connected, DC reachable by IP
- But `Resolve-DnsName dc01.domain.local` fails

**Possible causes:**

1. **NRPT rules not applied**
   - Check: `Get-DnsClientNrptRule`
   - Restart the service: `Restart-Service NetBirdMachine`

2. **DNS cache has old entries**
   - Clear cache: `Clear-DnsClientCache`

3. **DC DNS server not responding**
   - Test: `Test-NetConnection -ComputerName 192.168.100.20 -Port 53`

### Problem: Authentication fails on server

**Symptoms:**
- Client tries to connect
- Server log shows: "certificate validation failed" or "no peer auth method"

**Possible causes:**

1. **Certificate not from trusted CA**
   - Server must trust the CA that issued the client certificate
   - Check server CA configuration

2. **Certificate has wrong Subject Alternative Name (SAN)**
   - SAN must contain: `hostname.domain.local`
   - Check on client: Look at certificate details in certmgr.msc

3. **Certificate expired**
   - Check certificate validity dates
   - Request a new certificate from AD CS

---

## Getting Help

If you can't solve the problem:

1. **Collect diagnostic information:**
   - Service status: `Get-Service NetBirdMachine`
   - Last 100 lines of log: `Get-Content "C:\ProgramData\NetBird\machine-tunnel.log" -Tail 100`
   - Server logs: `docker logs ubuntu-management-1 --tail 100`

2. **Check the documentation:**
   - [Architecture Overview](ARCHITECTURE.md)
   - [Original NetBird Docs](https://docs.netbird.io/)

3. **Report issues:**
   - [GitHub Issues](https://github.com/obtFusi/netbird-fork/issues)

---

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - How the system works
- [NOTICE.md](../NOTICE.md) - License information
