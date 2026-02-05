# Architecture

This document describes the architecture of the NetBird Machine Tunnel fork, which extends NetBird with Windows pre-login VPN capabilities using Active Directory certificate authentication.

## Overview

The Machine Tunnel enables Windows machines to establish a VPN connection before user login, allowing domain authentication (Kerberos) to work for remote users. This is achieved through:

- **SYSTEM Service:** Runs as `NT AUTHORITY\SYSTEM` to operate before user login
- **mTLS Authentication:** Uses AD CS machine certificates instead of user credentials
- **Router-Peer Topology:** Domain Controllers remain untouched; dedicated Linux VMs route traffic

## Architecture Decisions

| Decision | Value | Rationale |
|----------|-------|-----------|
| **Scope** | NetBird Fork (Client + Server) | Server changes required for mTLS, AGPL publication accepted |
| **Windows Versions** | Windows 10/11 only | Server OS support not required |
| **Authentication** | Machine Certificates (mTLS) | Enterprise standard, AD CS integration |
| **DC Topology** | Router-Peer (HA: min. 2) | DCs remain unchanged, enterprise compliance |
| **DC Routing** | Return-Route (no NAT) | Clean auditing, no IP masking |
| **DNS** | NRPT (scoped) | Windows-native, GPO-deployable, selective |
| **ACL Enforcement** | Windows Firewall | WireGuard cannot filter by port |
| **Time Sync** | Public NTP before tunnel | Solves chicken-egg problem without tunnel dependency |

## Network Architecture

```
                              INTERNET
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                   NetBird Management Server                      │
│                   (management.example.com:443)                   │
│                   + mTLS Endpoint (:33074)                       │
└─────────────────────────────────────────────────────────────────┘
                                 │
               ┌─────────────────┴─────────────────┐
               │   WireGuard Mesh (100.64.0.0/10)  │
               │                                    │
       ┌───────┴───────┐                   ┌───────┴───────┐
       │               │                   │               │
       ▼               ▼                   ▼               │
  ┌────────┐     ┌────────┐          ┌──────────┐         │
  │ Win10  │     │ Win11  │          │ Router-  │         │
  │ Client │     │ Client │          │ Peer     │         │
  │100.64. │     │100.64. │          │100.64.   │         │
  │  x.x   │     │  x.x   │          │  x.x     │         │
  └────────┘     └────────┘          └────┬─────┘         │
       │              │                   │               │
       │              │                   │ ip_forward=1  │
       │              │                   ▼               │
       │              │         ┌─────────────────┐       │
       │              │         │   DC Network    │       │
       │              │         │ 192.168.x.0/24  │       │
       │              │         │                 │       │
       │              │         │  ┌───────────┐  │       │
       │              │         │  │   DC01    │  │       │
       │              │         │  └───────────┘  │       │
       │              │         │  ┌───────────┐  │       │
       │              │         │  │   DC02    │  │       │
       │              │         │  └───────────┘  │       │
       │              │         └─────────────────┘       │
       │              │                                   │
       └──────────────┴───────── WireGuard Tunnel ───────┘
```

## Router-Peer Concept

The Router-Peer is a dedicated Linux VM that routes traffic between the WireGuard mesh and the DC network. This design keeps Domain Controllers unchanged (no NetBird software installed).

**Responsibilities:**
- Runs NetBird peer with `ip_forward=1`
- Dual-NIC: One interface in DC network, one in client network
- Routes traffic between WireGuard tunnel and DC network
- No NAT - uses return routes for clean auditing

**High Availability:**
- Minimum 2 Router-Peers recommended
- Failover via WireGuard keepalive detection (~25s)
- Automatic recovery when peer comes back online

## Two-Phase Authentication

The Machine Tunnel uses a two-phase bootstrap process:

```
Phase 1: Setup Key (Bootstrap)
┌──────────┐     Setup Key      ┌──────────────┐
│  Client  │ ─────────────────► │  Management  │
│  (new)   │ ◄───────────────── │   Server     │
└──────────┘   Tunnel Config    └──────────────┘
     │
     │ Domain Join + Certificate Enrollment
     ▼
Phase 2: mTLS (Production)
┌──────────┐   Machine Cert     ┌──────────────┐
│  Client  │ ─────────────────► │  Management  │
│ (joined) │ ◄───────────────── │ Server:33074 │
└──────────┘   Tunnel Config    └──────────────┘
```

**Phase 1 (Setup Key):**
- One-time setup key with 24h TTL
- Establishes initial tunnel connection
- Allows domain join and certificate enrollment
- Setup key removed after successful bootstrap

**Phase 2 (mTLS):**
- Machine certificate from AD CS (SAN: `hostname.domain`)
- Connects to dedicated mTLS port (33074)
- No fallback to setup key - certificate required
- Issuer fingerprint validation for security

## Client Components

### Windows Service (`client/cmd/netbird-machine/`)

- Runs as `NT AUTHORITY\SYSTEM`
- Starts before user login (Session 0)
- SCM integration with automatic recovery
- Named pipe for IPC (future UI integration)

### Machine Tunnel (`client/internal/tunnel/`)

| Component | File | Purpose |
|-----------|------|---------|
| **Bootstrap** | `bootstrap.go` | Two-phase auth (Setup Key → mTLS) |
| **Machine** | `machine.go` | Main tunnel orchestration |
| **Health** | `health.go` | Connection monitoring, auto-reconnect |
| **Reconnect** | `reconnect.go` | Exponential backoff reconnection |
| **NRPT** | `nrpt_windows.go` | DNS Name Resolution Policy Table |
| **Firewall** | `firewall_windows.go` | Windows Firewall rules for DC access |
| **Event Log** | `eventlog_windows.go` | Windows Event Viewer integration |
| **Security** | `security_windows.go` | DPAPI encryption, secure config |
| **Peer Engine** | `peerengine.go` | Signal/Relay integration for NAT traversal |

### NRPT Integration

Name Resolution Policy Table routes DNS queries for AD domains through the tunnel:

```
*.corp.example.com  →  DC DNS Server (via tunnel)
*                   →  Normal DNS resolution
```

- Configured via Windows Registry
- Scoped by domain (not global)
- Hash-based key names for safe rollback

### Windows Firewall Rules

WireGuard `AllowedIPs` only filters by IP, not by port. The Machine Tunnel uses Windows Firewall for granular access control:

```
Allow: DC IPs on ports 53, 88, 389, 445, 636, etc.
Deny:  All other traffic on tunnel interface
```

- Rules grouped as "NetBird Machine Tunnel"
- Interface-specific (wg-nb-machine only)
- Automatic cleanup on service stop

## Server Components

### mTLS Endpoint

The management server exposes a dedicated mTLS endpoint:

| Port | Purpose | Client Auth |
|------|---------|-------------|
| 443 | Standard (Setup Key, SSO) | Optional |
| 33074 | mTLS-only (Machine Peers) | Required |

### Certificate Validation

Machine certificates are validated with:

1. **SAN DNSName:** Must match `hostname.domain` pattern
2. **AllowedDomains:** Per-account domain scoping (multi-tenant)
3. **Issuer Fingerprint:** Strong binding via certificate chain
4. **Expiry Check:** Certificate must be valid

## AD Ports Reference

Ports required for full AD functionality:

| Port | Protocol | Service |
|------|----------|---------|
| 53 | UDP + TCP | DNS |
| 88 | UDP + TCP | Kerberos |
| 123 | UDP | NTP |
| 135 | TCP | RPC Endpoint Mapper |
| 389 | UDP + TCP | LDAP |
| 445 | TCP | SMB (Sysvol, GPO) |
| 464 | UDP + TCP | Kerberos Password Change |
| 636 | TCP | LDAPS |
| 3268 | TCP | Global Catalog |
| 3269 | TCP | Global Catalog SSL |

**Note:** RPC dynamic ports should be restricted via GPO (e.g., 5000-5100) rather than using the default range (49152-65535).

## Health Monitoring

The health checker monitors tunnel status and triggers reconnection:

**Checks:**
- WireGuard interface up
- gRPC connection alive
- Last handshake recent (< 2 min)

**Behavior:**
- Check interval: 30 seconds
- Consecutive failures threshold: 3
- Auto-reconnect with exponential backoff

## Event Logging

Events are logged to Windows Event Viewer under source `NetBirdMachine`:

| Event ID | Description |
|----------|-------------|
| 1000 | Service Start |
| 1001 | Service Stop |
| 1100 | Tunnel Connected |
| 1101 | Tunnel Disconnected |
| 1200 | Auth Success |
| 1201 | Auth Failure |
| 1300 | ACL Hardened |
| 1301 | Setup Key Removed |
| 1400 | Config Error |

## Security Considerations

### Key Storage

- WireGuard private key encrypted with DPAPI (machine scope)
- Setup key removed after successful mTLS transition
- Certificates stored in Windows Certificate Store

### Network Isolation

- Clients cannot reach DC network without tunnel
- Router-Peer required for DC access
- Windows Firewall enforces port-level ACLs

### Certificate Security

- Short validity period (30 days recommended)
- Per-account AllowedDomains prevents cross-tenant access
- Issuer fingerprint prevents rogue CA attacks

## Related Documentation

- [NOTICE.md](../NOTICE.md) - AGPL-3.0 compliance and attribution
- [README.md](../README.md) - Quick start guide
