# NetBird Machine Tunnel Edition

**Complete NetBird VPN with pre-login Windows support for Active Directory environments.**

---

## About This Project

This is a fork of [NetBird](https://github.com/netbirdio/netbird) — an outstanding open-source WireGuard-based mesh VPN. We are deeply grateful to the NetBird team for building such an excellent foundation and for making it available under an open-source license.

**This fork is a complete, standalone NetBird deployment** — it includes all standard NetBird functionality (mesh networking, peer-to-peer connections, relay fallback, access controls, etc.) **plus** an additional feature called **Machine Tunnel** for Windows enterprise environments.

### What You Get

| Feature | Included |
|---------|----------|
| Full NetBird mesh VPN | ✅ Everything from upstream |
| Cross-platform clients | ✅ Windows, macOS, Linux, iOS, Android |
| Management Server | ✅ Complete with dashboard |
| Signal & Relay servers | ✅ NAT traversal included |
| **Machine Tunnel** | ✅ Pre-login VPN for Windows AD |

**When to use this fork:** If you need Windows machines to authenticate to Active Directory before user login (the "chicken and egg" problem), this fork solves it while giving you all of NetBird's capabilities.

**When to use upstream NetBird:** If you don't need pre-login Windows AD authentication, the [original NetBird](https://github.com/netbirdio/netbird) is an excellent choice with a larger community and more frequent updates.

<<<<<<< HEAD
---
=======
### Self-Host NetBird (Video)
[![Watch the video](https://img.youtube.com/vi/bZAgpT6nzaQ/0.jpg)](https://youtu.be/bZAgpT6nzaQ)
>>>>>>> upstream/main

## Table of Contents

- [Why This Fork Exists](#why-this-fork-exists)
- [How It Works](#how-it-works)
- [For End Users](#for-end-users)
  - [What You Need to Know](#what-you-need-to-know)
  - [Checking If It's Working](#checking-if-its-working)
  - [Common Questions](#common-questions)
- [For Administrators](#for-administrators)
  - [Prerequisites](#prerequisites)
  - [Architecture at a Glance](#architecture-at-a-glance)
  - [Installation](#installation)
  - [Configuration](#configuration)
  - [Security Checklist](#security-checklist)
  - [Verification](#verification)
  - [Troubleshooting Preview](#troubleshooting-preview)
- [What This Fork Adds to NetBird](#what-this-fork-adds-to-netbird)
- [Migration from Upstream NetBird](#migration-from-upstream-netbird)
- [Testing & Compatibility](#testing--compatibility)
- [Documentation](#documentation)
- [Building from Source](#building-from-source)
- [License](#license)

---

## Why This Fork Exists

### The Problem

If you've ever managed remote Windows users who need Active Directory, you've encountered this frustrating loop:

```
User wants to log in → Needs Kerberos ticket from Active Directory
                            ↓
              Needs network connection to Domain Controller
                            ↓
              Needs VPN to reach Domain Controller
                            ↓
              VPN needs user to be logged in first
                            ↓
                    🔄 Chicken and egg!
```

This hits hardest with:
- **New employees** receiving laptops remotely — they can't log in without AD
- **Remote workers** whose cached credentials expired
- **Password resets** that require immediate AD connectivity

### The Solution

Machine Tunnel breaks this cycle by connecting **before** any user logs in:

```
Computer boots → Machine Tunnel connects automatically (no user needed)
                            ↓
              Domain Controller is now reachable
                            ↓
              User logs in with Active Directory credentials
                            ↓
                    ✅ Problem solved!
```

The secret: instead of user credentials, we use the **computer's own certificate** (issued by Active Directory Certificate Services) to authenticate.

---

## How It Works

```
┌─────────────────────────────────────────────────────────────────────────┐
│                                                                          │
│   Remote Laptop                                           Data Center    │
│   ─────────────                                           ───────────    │
│                                                                          │
│   ┌─────────────┐         Encrypted WireGuard          ┌─────────────┐   │
│   │   Windows   │◄───────────Tunnel───────────────────►│   NetBird   │   │
│   │   Machine   │                                      │   Server    │   │
│   │   Tunnel    │                                      │             │   │
│   └──────┬──────┘                                      └─────────────┘   │
│          │                                                    │          │
│          │ Machine Certificate (mTLS)                         │          │
│          │                                                    │          │
│          ▼                                                    ▼          │
│   ┌─────────────┐                                      ┌─────────────┐   │
│   │   Windows   │      Kerberos, LDAP, DNS, SMB        │   Domain    │   │
│   │   Login     │◄────────────────────────────────────►│ Controller  │   │
│   │   Screen    │                                      │             │   │
│   └─────────────┘                                      └─────────────┘   │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

**What happens step by step:**

1. **Computer boots** — Windows starts the Machine Tunnel service automatically
2. **Certificate authentication** — The service proves identity using the computer's AD CS certificate
3. **Tunnel established** — Secure WireGuard connection to your NetBird server
4. **Domain Controller reachable** — AD traffic flows through the tunnel
5. **User logs in** — Kerberos works, even from a coffee shop halfway around the world

---

## For End Users

### What You Need to Know

**Good news:** If your IT department set this up, you don't need to do anything. Machine Tunnel runs automatically in the background, invisible to you.

**What changes for you:**

| Situation | Before | With Machine Tunnel |
|-----------|--------|---------------------|
| First login on new laptop (remote) | ❌ "Cannot contact domain controller" | ✅ Works normally |
| Password expired while working remotely | ❌ Stuck until you visit the office | ✅ Can change password |
| Cached credentials expired | ❌ Cannot log in | ✅ Fresh authentication from AD |

### Checking If It's Working

Want to verify? Open PowerShell and run:

```powershell
Get-Service NetBirdMachine
```

You should see:

```
Status   Name             DisplayName
------   ----             -----------
Running  NetBirdMachine   NetBird Machine Tunnel
```

### Common Questions

**Q: Do I need to start the VPN manually?**
No. Machine Tunnel starts automatically when your computer boots — even before you see the login screen.

**Q: Will this slow down my computer?**
No. The tunnel uses minimal resources and WireGuard is known for its performance.

**Q: What if I see "Cannot contact domain controller"?**
Contact your IT department. They can check the tunnel status and help you.

---

## For Administrators

### Prerequisites

| Requirement | Details |
|-------------|---------|
| **Windows** | Windows 10 or 11 (64-bit) |
| **Active Directory** | AD CS (Certificate Services) with machine certificate template |
| **Certificate** | Machine cert with SAN containing `hostname.domain` |
| **NetBird Server** | This fork with mTLS enabled |
| **Network** | Outbound access to ports 443 and 33074 |

### Architecture at a Glance

Machine Tunnel uses a **Router-Peer** design — your Domain Controllers remain completely untouched:

```
Windows Clients ──► WireGuard Mesh ──► Router-Peer VM ──► Domain Controllers
                                       (Linux, routing)
```

**Key benefits:**
- No software changes on Domain Controllers
- Easy to add redundancy (multiple Router-Peers)
- Clean audit trail (no NAT masking)

> **Full details:** See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for network diagrams, port requirements, and design decisions.

### Installation

#### Step 1: Prepare Your Server

Enable mTLS on the management server:

```bash
NETBIRD_MTLS_ENABLED=true
NETBIRD_MTLS_CA_FILE=/path/to/your-enterprise-ca.crt
NETBIRD_MTLS_PORT=33074
```

#### Step 2: Deploy to Windows Machines

```powershell
# Install the service
.\netbird-machine.exe install

# Verify it's running
Get-Service NetBirdMachine
```

Or use the installation script for automated deployment:

```powershell
.\scripts\install-netbird-machine.ps1 -ManagementURL "https://netbird.example.com" -SetupKey "your-key"
```

#### Step 3: Bootstrap New Machines

New machines use a two-phase process:

1. **Phase 1 (Setup Key):** Temporary key for initial connection and domain join
2. **Phase 2 (mTLS):** Machine certificate for permanent authentication

After the machine joins the domain and enrolls its certificate, the Setup Key is no longer needed.

> **Full details:** See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md#two-phase-authentication) for the complete bootstrap sequence.

### Configuration

Config file: `C:\ProgramData\NetBird\machine-config.yaml`

```yaml
management_url: "https://netbird.example.com:443"

# After certificate enrollment:
machine_cert_enabled: true

# Sensitive values are DPAPI-encrypted automatically:
# encrypted_private_key: <encrypted>
# encrypted_ssh_key: <encrypted>
```

**Security features built-in:**
- All secrets encrypted with Windows DPAPI (machine-bound)
- Config directory restricted to SYSTEM and Administrators
- Setup Key automatically removed after mTLS transition

### Security Checklist

Before going to production, verify these items:

| Check | Command |
|-------|---------|
| Service runs as SYSTEM | `Get-WmiObject Win32_Service -Filter "Name='NetBirdMachine'" \| Select StartName` |
| Config is encrypted | `Get-Content "C:\ProgramData\NetBird\machine-config.yaml"` — look for `encrypted_` prefix |
| Permissions restricted | `Get-Acl "C:\ProgramData\NetBird"` — only SYSTEM and Administrators |
| Certificate has correct SAN | Check certificate contains `hostname.domain` in SAN |
| mTLS port listening (server) | `ss -tlnp \| grep 33074` |

> **Full checklist:** See [docs/SECURITY-HARDENING.md](docs/SECURITY-HARDENING.md) for complete production hardening guide with verification scripts.

### Verification

After deployment, test connectivity:

```powershell
# 1. Service running?
Get-Service NetBirdMachine

# 2. Tunnel interface up?
Get-NetAdapter | Where-Object { $_.Name -like "wg*" }

# 3. Can reach Domain Controller?
Test-NetConnection -ComputerName dc01.yourdomain.local -Port 389
```

### Troubleshooting Preview

**Service won't start?**
```powershell
Get-Content "C:\ProgramData\NetBird\machine-tunnel.log" -Tail 50
```

**Tunnel up but can't reach DC?**
- Check Router-Peer is running
- Verify route exists: `route print | Select-String "192.168"`

**DNS not resolving AD names?**
```powershell
# Check NRPT rules
Get-DnsClientNrptRule

# Clear DNS cache
Clear-DnsClientCache
```

> **Full guide:** See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for step-by-step diagnostics and solutions.

---

## What This Fork Adds to NetBird

This fork includes **all standard NetBird features** plus the following additions for Windows AD environments:

| Feature | What It Does |
|---------|--------------|
| **Machine Tunnel Service** | Windows SYSTEM service that runs before user login |
| **mTLS Authentication** | Machine certificates from AD CS (in addition to Setup Keys, SSO, OIDC) |
| **NRPT DNS Routing** | Domain-specific DNS routing for AD traffic |
| **Windows Firewall Integration** | Port-level access control for Domain Controller traffic |
| **DPAPI Key Storage** | Hardware-bound encryption for sensitive configuration |
| **Dedicated mTLS Port (33074)** | Separate port for machine certificate authentication |

**Everything else works exactly like standard NetBird** — your existing NetBird clients, configurations, and workflows remain compatible.

---

## Migration from Upstream NetBird

### Can I Use Just the Machine Tunnel Feature?

**No.** The Machine Tunnel requires server-side changes that don't exist in upstream NetBird:
- mTLS authentication endpoint (port 33074)
- `RegisterMachinePeer` and `SyncMachinePeer` RPCs
- Per-account domain mapping for multi-tenant isolation

You need to deploy this fork's management server to use Machine Tunnel.

### Migrating an Existing NetBird Installation

If you're already running NetBird and want to add Machine Tunnel capability:

**Step 1: Replace the Management Server**
- Deploy this fork's management server
- Your existing configuration and peers will continue to work
- Standard NetBird clients are 100% compatible

**Step 2: Add Machine Tunnel to Windows Clients**
- Install `netbird-machine.exe` on machines that need pre-login VPN
- The standard `netbird` client can remain installed for post-login use
- Both services can coexist on the same machine

**Step 3: Configure mTLS**
- Set up AD CS certificate template
- Enable mTLS on the management server
- Enroll machine certificates on Windows clients

### Running Both Clients

You can run both the standard NetBird client and Machine Tunnel on the same Windows machine:

| Service | Runs As | When | Purpose |
|---------|---------|------|---------|
| `netbird` | Current User | After login | General VPN access |
| `netbird-machine` | SYSTEM | Before login | AD authentication |

They use separate configurations and don't interfere with each other.

---

## Testing & Compatibility

### What We Tested

The Machine Tunnel feature was tested in our lab environment:

| Component | Test Environment | What Was Verified |
|-----------|------------------|-------------------|
| **Windows Client** | Windows 11 (22H2) | Service lifecycle, DPAPI encryption, tunnel establishment |
| **Management Server** | Ubuntu 22.04 (Docker) | mTLS authentication, peer registration, multi-tenant isolation |
| **Domain Controller** | Windows Server 2022 | Certificate enrollment, Kerberos authentication |
| **Network** | Proxmox virtualization | Router-Peer routing, NRPT DNS, firewall rules |

**Tested scenarios:**
- Fresh machine bootstrap with Setup Key
- Domain join through tunnel
- Certificate enrollment from AD CS
- mTLS transition (Setup Key → Certificate)
- Service restart with key persistence
- Pre-login domain authentication

### What Comes from Upstream NetBird

The following components are inherited from NetBird and verified through their CI/CD:

| Component | Verified By |
|-----------|-------------|
| WireGuard tunnel implementation | NetBird CI |
| Peer-to-peer connectivity | NetBird CI |
| NAT traversal (STUN/TURN) | NetBird CI |
| Signal and Relay servers | NetBird CI |
| Management API (non-mTLS) | NetBird CI |
| Cross-platform clients (macOS, Linux, iOS, Android) | NetBird CI |
| Web dashboard | NetBird CI |

**Note:** We run NetBird's existing test suite as part of our CI, but we have not independently verified every upstream feature in our lab environment.

### Known Limitations

| Limitation | Details |
|------------|---------|
| Windows only | Machine Tunnel client supports Windows 10/11 only |
| AD CS required | Machine certificates must come from Active Directory Certificate Services |
| No CRL checking | Certificate revocation checking is not yet implemented |

---

## Documentation

| Document | What's Inside |
|----------|---------------|
| [ARCHITECTURE.md](docs/ARCHITECTURE.md) | Network topology, component design, bootstrap sequence |
| [SECURITY-HARDENING.md](docs/SECURITY-HARDENING.md) | DPAPI, file permissions, mTLS setup, production checklist |
| [TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) | Health checks, diagnostics, common problems with solutions |
| [CHANGELOG.md](CHANGELOG.md) | All changes from upstream with file listings |
| [NOTICE.md](NOTICE.md) | License attribution and modification summary |

---

## Building from Source

**Requirements:**
- Go 1.21+
- For Windows builds from Linux: MinGW-w64

**Build:**

```bash
# Windows binary
GOOS=windows GOARCH=amd64 go build -o netbird-machine.exe ./client/cmd/netbird-machine

# With CGO (for DPAPI)
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 go build -o netbird-machine.exe ./client/cmd/netbird-machine
```

**Test:**

```bash
go test ./client/internal/tunnel/...
```

---

## License

Licensed under **GNU Affero General Public License v3.0 (AGPL-3.0)** — same as NetBird.

- You can use, modify, and distribute this software
- If you modify and deploy it, you must share your source code
- Full text: [LICENSE](LICENSE)

### Attribution

Based on [NetBird](https://github.com/netbirdio/netbird) by the NetBird Authors.

See [CHANGELOG.md](CHANGELOG.md) for technical details of our modifications.
See [NOTICE.md](NOTICE.md) for modification summary and attribution.

---

## Contributing

For general NetBird improvements, please contribute to the [upstream project](https://github.com/netbirdio/netbird).

For Machine Tunnel-specific issues, open an issue here.

---

*Thank you to the NetBird team for creating such excellent open-source software.*
