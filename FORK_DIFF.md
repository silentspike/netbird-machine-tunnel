# Fork Contribution: Machine Tunnel

This document summarizes what this fork adds on top of upstream NetBird. It is
written for reviewers who need to understand the fork-specific contribution
quickly without reading the full upstream-sized repository.

Baseline used for review: upstream NetBird `v0.69.0`.

## Summary

This fork keeps the full NetBird product surface and adds a Windows Machine
Tunnel for Active Directory environments. The Machine Tunnel authenticates a
Windows machine before user login, uses AD CS machine certificates for mTLS, and
adds server-side registration/sync paths for machine peers.

The core additions are:

- Windows pre-login service entry point.
- Machine bootstrap from setup key to certificate-backed mTLS.
- Windows certificate-store discovery and CNG signing for non-exportable
  machine certificates.
- Dedicated management mTLS server for Machine Tunnel RPCs.
- Machine-peer registration, sync, routes, and status RPCs.
- DNS, firewall, routing, health, and reconnect support for the machine tunnel.
- Public-release hardening around CI, secrets, branch protection, and release
  artifacts.

## Client Additions

### Windows Service

- `client/cmd/netbird-machine/`
  - Windows service entry point.
  - Service configuration parsing.
  - Service Control Manager integration.
  - Installation helper for the Machine Tunnel service.

### Machine Tunnel Orchestration

- `client/internal/tunnel/`
  - Setup-key bootstrap.
  - Certificate enrollment transition.
  - Domain join support.
  - Windows Event Log integration.
  - Windows Firewall and DC reachability handling.
  - Tunnel health and reconnect logic.
  - Windows interface handling.
  - Machine lifecycle orchestration.
  - NRPT and trust handling.
  - Peer engine integration for Signal/Relay reuse.

### Machine Certificate Authentication

- `client/internal/auth/cert_discovery_windows.go`
  - Opens the Windows `LocalMachine\My` certificate store.
  - Enumerates and selects machine certificates.
  - Supports thumbprint lookup and AD CS template metadata.
  - Acquires private-key handles via CryptoAPI/CNG.

- `client/internal/auth/wincert_signer.go`
  - Implements `crypto.Signer` for Windows certificate-store keys.
  - Uses `NCryptSignHash` without exporting private keys.
  - Supports RSA and ECDSA signing paths.

- `client/internal/auth/mtls_client.go`
  - Provides the mTLS management client for Machine Tunnel RPCs.

- `client/internal/auth/wincert_signer_other.go` and
  `client/internal/auth/cert_discovery_other.go`
  - Keep non-Windows builds explicit and safe.

### Windows Networking Support

- `client/internal/dns/nrpt_windows.go`
- `client/internal/firewall/`
- `client/internal/iface/discovery_windows.go`
- `client/internal/ntp/`

These packages support Windows DNS policy, Domain Controller reachability,
firewall configuration, interface discovery, and time synchronization needed for
pre-login domain authentication.

## Management Server Additions

### Dedicated mTLS Server

- `management/internals/server/mtls_server.go`
  - Dedicated mTLS-only gRPC server.
  - Defaults to port `33074`.
  - Uses `tls.RequireAndVerifyClientCert`.

- `management/internals/server/server.go`
  - Starts and stops the dedicated mTLS server with the main management server
    lifecycle.

- `management/internals/server/mtls_auth.go`
  - Extracts and validates machine identities from client certificates.
  - Maps certificate DNS names/domains to accounts.
  - Enforces account-level issuer and domain constraints.
  - Provides unary and stream interceptors for Machine Tunnel RPCs.

### Machine Tunnel RPCs

- `management/internals/shared/grpc/machine_tunnel.go`
  - Implements machine peer registration.
  - Implements machine peer sync.
  - Provides route and status handling for machine peers.

- `shared/management/proto/management.proto`
- `shared/management/proto/management.pb.go`
- `shared/management/proto/management_grpc.pb.go`
  - Add generated client/server code for:
    - `RegisterMachinePeer`
    - `SyncMachinePeer`
    - `GetMachineRoutes`
    - `ReportMachineStatus`

### Peer and Audit Integration

- `management/server/peer.go`
  - Adds machine-peer registration behavior for mTLS-authenticated peers.
  - Adds DNS-label handling for machine certificate identities.

- `management/server/peer/peer.go`
  - Adds peer metadata for machine authentication state.

- `management/server/activity/codes.go`
  - Adds Machine Tunnel activity code(s), including mTLS peer registration.

## Tests

Representative fork-specific tests include:

- `client/internal/auth/wincert_pss_test.go`
- `client/internal/tunnel/bootstrap_test.go`
- `client/internal/tunnel/certenroll_test.go`
- `client/internal/tunnel/domainjoin_test.go`
- `client/internal/tunnel/health_test.go`
- `client/internal/tunnel/mtls_test.go`
- `client/internal/tunnel/peerengine_test.go`
- `client/internal/tunnel/reconnect_test.go`
- `client/internal/tunnel/trust_test.go`
- `management/internals/server/mtls_auth_test.go`
- `management/internals/shared/grpc/machine_tunnel_test.go`
- `management/internals/shared/mtls/dnslabel_test.go`
- `management/internals/shared/mtls/identity_test.go`
- `management/server/geolocation/utils_test.go`
- `management/server/idp/util_test.go`

The release process also uses GitHub Actions and lab validation to cover
Windows, Linux, Darwin, FreeBSD, Mobile, Wasm, release artifacts, secret
scanning, generated proto checks, and dependency/license checks.

## Public-Readiness Work

This fork also adds or modifies public-governance files that are not Machine
Tunnel runtime code but are important for operating the fork as a public
repository:

- `.github/CODEOWNERS`
- `.github/dependabot.yml`
- `.github/workflows/codeql.yml`
- `.github/workflows/secret-scan.yml`
- `.github/workflows/check-proto-generate.yml`
- `.github/workflows/e2e-tunnel.yml`
- `.github/workflows/release.yml`
- `.gitleaks.toml`
- `NOTICE.md`
- `SECURITY.md`
- `CHANGELOG.md`
- `llms.txt`

## Fork Markers

Some fork-specific server changes are marked with comments containing
`MACHINE-TUNNEL-FORK`, but this is not a complete index of all fork changes.
Use path review and upstream comparison as the source of truth.

Useful review commands:

```bash
git diff --name-status v0.69.0...HEAD -- \
  client/cmd/netbird-machine \
  client/internal/tunnel \
  client/internal/auth \
  management/internals/server \
  management/internals/shared/grpc \
  management/internals/shared/mtls \
  management/server/peer.go \
  management/server/activity/codes.go \
  shared/management/proto

rg -n "MACHINE-TUNNEL-FORK|RegisterMachinePeer|SyncMachinePeer|MTLSServerPort|WinCertSigner|NCryptSignHash" \
  client management shared/management/proto
```

## Known Security Limitation

Certificate revocation list checking is not implemented yet. The current
mitigations are short certificate lifetimes, explicit account/domain mapping,
issuer fingerprint validation, and operational certificate rotation. See
`README.md` and `SECURITY.md` for the public security posture.
