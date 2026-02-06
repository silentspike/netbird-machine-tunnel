# Notice

This project is a fork of [NetBird](https://github.com/netbirdio/netbird).

## Licensing

NetBird uses a **dual-license** structure:

| Component | License | Directories |
|-----------|---------|-------------|
| Client, Shared Libraries | **BSD-3-Clause** | `client/`, `shared/`, root |
| Management Server | **AGPL-3.0** | `management/` |
| Signal Server | **AGPL-3.0** | `signal/` |
| Relay Server | **AGPL-3.0** | `relay/` |

This fork inherits and follows the same dual-license structure. See [LICENSE](LICENSE) for the BSD-3-Clause text, and the respective `LICENSE` files in `management/`, `signal/`, and `relay/` for AGPL-3.0.

## Original Project

- **Project:** NetBird
- **Repository:** https://github.com/netbirdio/netbird
- **Copyright:** NetBird Authors

## Fork Information

- **Fork Repository:** https://github.com/obtFusi/netbird-fork
- **Fork Purpose:** Machine Tunnel for Windows Pre-Login VPN with AD Certificate Authentication

## Modifications

This fork includes the following modifications from upstream NetBird:

### Machine Tunnel Feature (client/internal/tunnel/)
- **mTLS Authentication:** Machine certificate authentication using Windows AD CS
- **NRPT Integration:** DNS Name Resolution Policy Table for AD domain routing
- **Windows Firewall Rules:** Granular port-based ACLs for Domain Controller access
- **Health Monitoring:** Tunnel health checks with auto-reconnect
- **Windows Event Log:** Service events logged to Windows Event Viewer
- **Crash Recovery:** Automatic reconnection with exponential backoff
- **DPAPI Security:** Secure key storage using Windows Data Protection API

### Windows Service (client/cmd/netbird-machine/)
- **SYSTEM Service:** Runs as NT AUTHORITY\SYSTEM for pre-login operation
- **SCM Integration:** Windows Service Control Manager with recovery actions
- **Bootstrap Flow:** Two-phase authentication (Setup Key, then mTLS)

### Server Extensions (management/server/)
- **Machine Peer Registration:** RegisterMachinePeer RPC for machine authentication
- **mTLS Validation:** Certificate validation with AllowedDomains per account
- **Issuer Fingerprint:** Strong binding via certificate chain validation

## Source Code Availability

As required by AGPL-3.0 (for the server components), the complete source code for this fork is available at:

https://github.com/obtFusi/netbird-fork

## Original Authors

NetBird was created by the NetBird team and contributors:
https://github.com/netbirdio/netbird/graphs/contributors
