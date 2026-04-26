# Contributing to NetBird Machine Tunnel Edition

Thank you for your interest in contributing to this project.

This is a fork of [NetBird](https://github.com/netbirdio/netbird) that adds a Machine Tunnel feature for Windows pre-login VPN. Contributions to both the Machine Tunnel feature and general improvements are welcome.

## Where to Report Issues

| Type | Where |
|------|-------|
| Machine Tunnel bugs/features | [This repository](https://github.com/silentspike/netbird-machine-tunnel/issues) |
| Standard NetBird bugs/features | [Upstream NetBird](https://github.com/netbirdio/netbird/issues) |
| Security vulnerabilities | See [SECURITY.md](SECURITY.md) |
| General questions | [Upstream NetBird Discussions](https://github.com/netbirdio/netbird/discussions) |

If you are unsure whether an issue belongs here or upstream, open it here and we will redirect if needed.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How to Contribute

### 1. Fork and Clone

```bash
git clone https://github.com/<your_username>/netbird-machine-tunnel.git
cd netbird-machine-tunnel
git remote add upstream https://github.com/silentspike/netbird-machine-tunnel.git
```

### 2. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

### 3. Make Changes

- Follow existing code patterns and style
- Add tests for new functionality
- Update documentation for user-visible changes

### 4. Test

```bash
# Run all tests
go test ./...

# Run Machine Tunnel specific tests
go test ./client/internal/tunnel/...
go test ./management/internals/shared/grpc/... -run MachineTunnel

# Run linter
golangci-lint run

# Build Windows binary
GOOS=windows GOARCH=amd64 go build -o bin/netbird-machine.exe ./client/cmd/netbird-machine
```

### 5. Submit a Pull Request

Open a PR against the `main` branch with a clear description of what changed and why.

## Directory Structure

This fork follows the standard NetBird directory structure with additions:

### Fork-Specific Directories

| Path | Purpose |
|------|---------|
| `client/cmd/netbird-machine/` | Machine Tunnel Windows service |
| `client/internal/tunnel/` | Machine Tunnel core logic |
| `management/internals/server/mtls_*` | Server-side mTLS authentication |
| `management/internals/shared/mtls/` | mTLS identity extraction |
| `scripts/` | Deployment and testing scripts |
| `docs/` | Architecture and security documentation |

### Standard NetBird Directories

| Path | Purpose |
|------|---------|
| `client/` | NetBird client agent |
| `management/` | Management server |
| `signal/` | Signal server |
| `relay/` | Relay server |
| `shared/` | Shared protobuf definitions and utilities |

For the complete upstream directory structure, see the [NetBird CONTRIBUTING.md](https://github.com/netbirdio/netbird/blob/main/CONTRIBUTING.md).

## Understanding Fork-Specific Code

Code added or modified by this fork is marked with comments:

```go
// === MACHINE-TUNNEL-FORK START ===
// Description of what this block does
yourCode()
// === MACHINE-TUNNEL-FORK END ===
```

These markers serve two purposes:
1. Help contributors identify fork-specific code vs upstream code
2. Assist with merge conflict resolution during upstream syncs

**When modifying upstream files**, always use these markers around your changes.

## Development Requirements

- **Go 1.21+** (see `go.mod` for exact version)
- **golangci-lint** (matching version in `.github/workflows/golangci-lint.yml`)
- **protoc** with `protoc-gen-go` and `protoc-gen-go-grpc` (for proto changes)

### Windows-Specific Development

Machine Tunnel features require testing on Windows:

- Windows 10/11 or Windows Server 2019+
- Active Directory environment (for certificate tests)
- Administrator privileges (for service installation)

## Pull Request Checklist

Before submitting a PR, verify:

- [ ] Code compiles without errors (`go build ./...`)
- [ ] All tests pass (`go test ./...`)
- [ ] Linter passes (`golangci-lint run`)
- [ ] Windows cross-compilation succeeds (`GOOS=windows GOARCH=amd64 go build ./client/cmd/netbird-machine`)
- [ ] New public functions have tests
- [ ] User-visible changes are documented in the PR description
- [ ] CHANGELOG.md is updated for features and bug fixes
- [ ] Fork-specific code uses `MACHINE-TUNNEL-FORK` markers in upstream files
- [ ] No secrets, credentials, or internal IPs in the code

## Upstream Synchronization

This fork is regularly synchronized with upstream NetBird via an automated workflow. The sync runs weekly and creates a PR for manual review.

When contributing, keep in mind:
- Avoid modifying upstream files when possible
- When modifications are necessary, use `MACHINE-TUNNEL-FORK` markers
- Activity IDs 150+ are reserved for this fork (upstream uses 0-149)
- Keep fork-specific code isolated in dedicated files/packages

## Coding Guidelines

### Go Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Wrap errors with context: `fmt.Errorf("operation description: %w", err)`
- Use `context.Context` for cancellation
- Keep interfaces small and focused

### Windows-Specific Code

- Use `_windows.go` suffix for Windows-only files
- Use `golang.org/x/sys/windows` for Windows API calls
- Store sensitive data with DPAPI, not plaintext
- Use registry paths under `HKLM:\SOFTWARE\NetBird\Machine`

### Security

- Validate certificates via SAN DNSName, not CN
- Use `VerifiedChains` for issuer fingerprinting
- Never log private keys, certificates, or credentials
- Scope firewall rules to the `wg-nb-machine` interface

## Contributor License Agreement

By submitting a pull request, you agree to the terms of the [Contributor License Agreement](CONTRIBUTOR_LICENSE_AGREEMENT.md).

## Need Help?

- Read the [Architecture Documentation](docs/ARCHITECTURE.md) to understand the system
- Check the [Troubleshooting Guide](docs/TROUBLESHOOTING.md) for common issues
- Review existing [issues](https://github.com/silentspike/netbird-machine-tunnel/issues) for context
- Open a discussion if you need guidance on an approach
