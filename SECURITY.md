# Security Policy

NetBird Machine Tunnel Edition handles enterprise VPN connectivity with Active Directory certificate authentication. Security is a top priority.

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest release | Yes |
| Previous minor | Security fixes only |
| Older versions | No |

We recommend always running the latest release. Security patches are backported to the previous minor version on a best-effort basis.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

### How to Report

1. **GitHub Security Advisories (preferred):** Use the [private vulnerability reporting](https://github.com/silentspike/netbird-machine-tunnel/security/advisories/new) feature on this repository.

2. **Maintainer contact:** If private vulnerability reporting is unavailable, contact the repository maintainers listed in [CODEOWNERS](.github/CODEOWNERS).

### What to Include

- Description of the vulnerability
- Steps to reproduce (proof of concept if possible)
- Affected versions
- Potential impact assessment
- Suggested fix (if you have one)

### What to Expect

| Timeframe | Action |
|-----------|--------|
| 48 hours | Acknowledgment of your report |
| 7 days | Initial assessment and severity classification |
| 30 days | Fix developed and tested (for confirmed vulnerabilities) |
| 90 days | Public disclosure (coordinated with reporter) |

We follow [coordinated vulnerability disclosure](https://en.wikipedia.org/wiki/Coordinated_vulnerability_disclosure). We will not take legal action against researchers who follow responsible disclosure practices.

## Scope

### In Scope

- Machine Tunnel client (`netbird-machine.exe`)
- mTLS authentication and certificate validation
- DPAPI key storage (`security_windows.go`)
- WireGuard key management and persistence
- Server-side mTLS endpoint (`management/internals/server/mtls_*`)
- NRPT DNS configuration
- Windows Firewall rule management
- Bootstrap sequence (setup key handling)
- Any code in the `MACHINE-TUNNEL-FORK` marked sections

### Out of Scope

- Standard NetBird functionality (report to [upstream](https://github.com/netbirdio/netbird/security))
- Vulnerabilities in third-party dependencies (report to the respective project, but let us know so we can update)
- Issues requiring physical access to the machine
- Social engineering attacks
- Denial of service attacks against lab/test infrastructure

## Security Architecture

For details on the security design of the Machine Tunnel feature, see:

- [Security Hardening Guide](docs/SECURITY-HARDENING.md)
- [Architecture Documentation](docs/ARCHITECTURE.md)

### Key Security Properties

- **Machine certificates** validated via SAN DNSName (not CN), with issuer fingerprint verification through `VerifiedChains`
- **Private keys** protected with Windows DPAPI (machine scope, additional entropy)
- **Setup keys** encrypted at rest, revoked after bootstrap
- **Port-level ACLs** enforced via Windows Firewall (interface-specific), not WireGuard AllowedIPs
- **NRPT rules** scoped with hash-based registry keys for safe rollback
- **Multi-tenant isolation** through per-account AllowedDomains scoping

## Known Security Limitations

### Certificate Revocation Checking

Machine Tunnel mTLS currently validates the presented machine certificate chain,
SAN DNSName, issuer fingerprint, and per-account AllowedDomains mapping, but it
does not perform CRL or OCSP revocation checks. This limitation applies to the
Machine Tunnel certificate authentication path on the dedicated mTLS endpoint;
it does not change upstream NetBird authentication behavior.

Until revocation checking is implemented, an already-issued machine certificate
can continue to authenticate until the certificate expires unless operators
remove the matching account/domain mapping, change the accepted issuer
fingerprint, or rotate the machine certificate population. Production
deployments should use short machine-certificate lifetimes, routine certificate
rotation, tightly scoped AD CS issuance permissions, per-account AllowedDomains,
and issuer fingerprint constraints.

## Upstream Security

This fork inherits the security properties of [NetBird](https://github.com/netbirdio/netbird). For vulnerabilities in upstream NetBird components, please report to `security@netbird.io` and notify us so we can track the fix.

## Acknowledgments

We appreciate the security research community's efforts in keeping this project safe. Confirmed vulnerability reporters will be credited in the release notes (unless they prefer to remain anonymous).
