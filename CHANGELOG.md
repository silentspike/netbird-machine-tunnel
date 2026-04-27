# Changelog

All notable changes to the NetBird Machine Tunnel fork will be documented in this file.

This project is a fork of [NetBird](https://github.com/netbirdio/netbird). The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

---

## [Unreleased] - 2026-04-24

### Changed

- Synced the fork base to upstream NetBird `v0.69.0`.
- Updated Machine Tunnel setup-key bootstrap to the new upstream management client API (`Login(sysInfo, sshKey, dnsLabels)` and `Register(setupKey, jwtToken, sysInfo, sshKey, dnsLabels)`).
- Kept Machine Tunnel peer connections deterministic by explicitly leaving upstream `PortForwardManager` and `MetricsRecorder` disabled for the pre-login PeerEngine path.
- Updated release workflow and GoReleaser configs to generate SPDX JSON SBOMs for release archives and Linux packages through Syft.
- Fixed the upstream-sync workflow to merge the selected upstream release tag instead of `upstream/main`.

### Fixed

- Hardened fork-owned CodeQL follow-up findings before public release: Machine Tunnel DNS labels now use a deterministic HMAC-SHA256 64-bit suffix, geolocation archive extraction writes only expected files and skips traversal entries, and the branch-local gRPC sync-limit parser now uses bounded `strconv.ParseInt(..., 32)` parsing.

### Added

- Upstream `v0.69.0` client and proxy features, including port-forwarding, client connection metrics, reverse-proxy L4/CrowdSec-related code, IDP migration helpers, and expanded management API tests.

### Notes

- CrowdSec code is included from upstream but is not enabled in the Machine Tunnel lab profile.
- Machine Tunnel public release validation has passed local, CI, and lab gates on the upgrade branch; post-merge main smoke, release-candidate artifact validation, and final go-live approval remain required before publishing a public tag.

## [1.0.0-machine-tunnel] - 2026-02-05

This is the first release of the Machine Tunnel feature. It enables Windows machines to establish a VPN connection before user login, allowing domain authentication (Kerberos) to work for remote users.

### Added - Client Components

#### New Package: `client/internal/tunnel/`

Complete Machine Tunnel implementation for Windows:

| File | Lines | Purpose |
|------|-------|---------|
| `machine.go` | ~1200 | Main tunnel orchestration, state machine, connection management |
| `bootstrap.go` | ~450 | Two-phase authentication: Setup Key → mTLS transition |
| `peerengine.go` | ~450 | Signal/Relay integration for NAT traversal |
| `security_windows.go` | ~450 | DPAPI encryption, secure config storage |
| `health.go` | ~350 | Connection health monitoring, auto-reconnect |
| `reconnect.go` | ~200 | Exponential backoff reconnection logic |
| `nrpt_windows.go` | ~200 | Windows NRPT (Name Resolution Policy Table) for DNS routing |
| `firewall_windows.go` | ~250 | Windows Firewall rules for DC traffic |
| `interface_windows.go` | ~250 | WireGuard interface management |
| `eventlog_windows.go` | ~80 | Windows Event Log integration |
| `certenroll.go` | ~350 | AD CS certificate enrollment helpers |
| `domainjoin.go` | ~300 | Domain join via tunnel helpers |
| `trust_windows.go` | ~300 | CA trust chain validation |

#### New Package: `client/cmd/netbird-machine/`

Windows service entry point:

| File | Lines | Purpose |
|------|-------|---------|
| `main.go` | ~100 | CLI entry point, commands registration |
| `service_windows.go` | ~300 | Windows Service implementation (SCM integration) |
| `install_windows.go` | ~200 | Service installation and uninstallation |

#### New Scripts: `scripts/`

PowerShell scripts for deployment and testing:

| Script | Purpose |
|--------|---------|
| `install-netbird-machine.ps1` | Automated installation with config generation |
| `reset-netbird-machine.ps1` | Clean reset for testing (removes service, config, interface) |
| `reinstall-and-test.ps1` | Reinstall and verify functionality |
| `verify-config-hardening.ps1` | Security verification checks |
| `verify-nrpt-cleanup.ps1` | NRPT cleanup verification |
| `bootstrap-new-client.ps1` | Bootstrap new client with Setup Key |

#### New Scripts: `scripts/tests/`

E2E test scripts:

| Script | Purpose |
|--------|---------|
| `Test-TunnelEstablishment.ps1` | Comprehensive tunnel connectivity tests |
| `Test-Baseline.ps1` | System baseline verification |
| `Test-GoNoGoMatrix.ps1` | Pre-flight checks before E2E tests |
| `Test-DomainJoinViaTunnel.ps1` | Domain join over tunnel test |
| `Invoke-E2ETestCycle.ps1` | Complete E2E test orchestration |

### Added - Server Components

#### New Package: `management/internals/server/`

mTLS authentication for machine peers:

| File | Lines | Purpose |
|------|-------|---------|
| `mtls_server.go` | ~200 | Dedicated mTLS gRPC server on port 33074 |
| `mtls_auth.go` | ~600 | Certificate validation, domain mapping, issuer checks |
| `mtls_auth_test.go` | ~500 | Comprehensive mTLS authentication tests |

#### New Package: `management/internals/shared/mtls/`

Shared mTLS utilities:

| File | Purpose |
|------|---------|
| `identity.go` | mTLS identity extraction from certificates |
| `dnslabel.go` | DNS label generation for unique peer naming |
| `dnslabel_test.go` | DNS label generation tests |

#### New Package: `management/internals/shared/grpc/`

Machine Tunnel RPC handlers:

| File | Lines | Purpose |
|------|-------|---------|
| `machine_tunnel.go` | ~500 | RegisterMachinePeer, SyncMachinePeer, GetMachineRoutes handlers |
| `machine_tunnel_test.go` | ~700 | Handler tests with mocked dependencies |

#### Modified: `management/internals/server/server.go`

- Added mTLS server initialization
- Registered Machine Tunnel handlers on mTLS port
- Added `mTLSRequiredMethods` enforcement

#### Modified: `management/server/peer.go`

- Added `AddMachinePeer` function for certificate-based registration
- Extended peer metadata for machine peers

### Added - Documentation

| File | Purpose |
|------|---------|
| `docs/ARCHITECTURE.md` | System architecture and design decisions |
| `docs/TROUBLESHOOTING.md` | Diagnostic commands and common problems |
| `docs/SECURITY-HARDENING.md` | Security configuration guide |
| `docs/ADR-001-mTLS-Port-Strategy.md` | Architecture Decision Record for port split |
| `docs/ADR-002-CNG-Signer-Interface.md` | Architecture Decision Record for CNG integration |
| `NOTICE.md` | AGPL-3.0 compliance and attribution |

### Added - CI/CD

| File | Purpose |
|------|---------|
| `.github/workflows/e2e-tunnel.yml` | E2E test workflow (manual trigger) |
| `.github/workflows/golang-test-windows.yml` | Windows-specific test workflow |

### Changed - Configuration

#### New Configuration Options

```yaml
# Machine Tunnel specific settings in machine-config.yaml
management_url: "https://server:443"
encrypted_setup_key: "<DPAPI-encrypted>"     # Bootstrap key (removed after mTLS transition)
encrypted_private_key: "<DPAPI-encrypted>"   # WireGuard private key
encrypted_ssh_key: "<DPAPI-encrypted>"       # SSH key for peer connections
machine_cert_enabled: false                   # Enable mTLS with machine certificate
key_version: 1                                # Config version for future migrations
```

#### New Server Configuration

| Environment Variable | Purpose |
|---------------------|---------|
| `NETBIRD_MTLS_ENABLED` | Enable mTLS authentication (default: false) |
| `NETBIRD_MTLS_CA_FILE` | Path to CA certificate for client validation |
| `NETBIRD_MTLS_PORT` | Port for mTLS-only connections (default: 33074) |
| `NETBIRD_MTLS_DOMAIN_MAPPING` | JSON mapping of domains to account IDs |
| `NETBIRD_MTLS_ALLOWED_ISSUERS` | JSON mapping of account IDs to allowed CA fingerprints |

### Changed - Network Architecture

#### New Port

| Port | Protocol | Purpose |
|------|----------|---------|
| 33074 | TCP/gRPC | mTLS-only Machine Tunnel authentication |

Standard NetBird ports (443, 33073) remain unchanged and functional.

#### Router-Peer Topology

The Machine Tunnel uses a Router-Peer architecture where Domain Controllers remain untouched:

```
Windows Client ──► WireGuard Mesh ──► Router-Peer VM ──► Domain Controller
                   (100.64.0.0/10)    (Linux, ip_forward=1)  (192.168.x.0/24)
```

### Security Features

#### Client-Side Security

| Feature | Implementation |
|---------|----------------|
| **DPAPI Encryption** | All secrets encrypted with machine-scope DPAPI |
| **File Permissions** | Config directory restricted to SYSTEM + Administrators |
| **Service Account** | Runs as LocalSystem for pre-login operation |
| **Certificate Binding** | Uses AD CS machine certificates with SAN validation |

#### Server-Side Security

| Feature | Implementation |
|---------|----------------|
| **mTLS Enforcement** | Machine-specific RPCs require client certificates |
| **Domain Isolation** | Per-account AllowedDomains prevents cross-tenant access |
| **Issuer Validation** | SHA256 fingerprint check of certificate issuer CA |
| **Multi-Account Detection** | Rejects certificate SANs that span multiple accounts |

### Technical Details

#### gRPC Methods (mTLS-only)

These methods are only accessible via mTLS on port 33074:

| Method | Purpose |
|--------|---------|
| `RegisterMachinePeer` | Register machine with certificate identity |
| `SyncMachinePeer` | Sync configuration and peer list |
| `GetMachineRoutes` | Get routes for DC network access |
| `ReportMachineStatus` | Report health status to management |

#### Windows Event Log

Events logged under source `NetBirdMachine`:

| Event ID | Description |
|----------|-------------|
| 1000 | Service Start |
| 1001 | Service Stop |
| 1100 | Tunnel Connected |
| 1101 | Tunnel Disconnected |
| 1200 | Authentication Success |
| 1201 | Authentication Failure |
| 1300 | ACL Hardened |
| 1301 | Setup Key Removed |
| 1400 | Configuration Error |

### Dependencies

No new external dependencies added. Uses existing NetBird dependencies:
- `golang.zx2c4.com/wireguard` - WireGuard implementation
- `google.golang.org/grpc` - gRPC framework
- `github.com/pion/ice/v4` - NAT traversal
- `golang.org/x/sys/windows` - Windows API access

### Compatibility

- **Base Version:** NetBird main branch (January 2026)
- **Go Version:** 1.21+
- **Windows:** Windows 10/11 (64-bit)
- **Server:** Linux (tested on Ubuntu 22.04)

### Known Limitations

1. **Windows Only:** Machine Tunnel client currently supports Windows only
2. **mTLS Required:** After bootstrap, machine certificate is mandatory (no password fallback)
3. **AD CS Required:** Certificate enrollment requires Active Directory Certificate Services

---

## Differences from Upstream NetBird

| Aspect | Upstream NetBird | This Fork |
|--------|------------------|-----------|
| **Authentication** | Setup keys, SSO, OIDC | + Machine certificates (mTLS) |
| **Tunnel Start** | After user login | Before login (pre-logon) |
| **Service Type** | User-space service | Windows SYSTEM service |
| **DNS Routing** | Global or per-interface | NRPT (domain-specific) |
| **Firewall** | WireGuard AllowedIPs | + Windows Firewall rules |
| **Key Storage** | Plaintext config | DPAPI encrypted |
| **Server Ports** | 443, 33073 | + 33074 (mTLS-only) |

---

## Migration from Upstream

If migrating from standard NetBird to this fork:

1. Standard NetBird clients continue to work unchanged
2. Machine Tunnel requires separate installation (`netbird-machine.exe`)
3. Both can coexist on the same management server
4. Machine Tunnel uses dedicated port 33074

---

## License

This fork maintains the same AGPL-3.0 license as upstream NetBird.
See [NOTICE.md](NOTICE.md) for attribution details.
