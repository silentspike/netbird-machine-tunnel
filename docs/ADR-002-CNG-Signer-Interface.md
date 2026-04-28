# ADR-002: Windows CNG crypto.Signer Interface

**Status:** Implemented
**Date:** 2026-01-20
**Issue:** T-1.1 (Windows CNG crypto.Signer Spike)

> 2026-04-27 update: the Windows certificate-store signer and certificate
> discovery path are implemented. See `client/internal/auth/wincert_signer.go`,
> `client/internal/auth/cert_discovery_windows.go`, and the non-Windows stubs.

## Context

Machine Tunnel needs to use Windows machine certificates stored in the Windows Certificate Store for mTLS authentication. Go's standard `crypto/tls` expects a `crypto.Signer` interface, but Windows certificates use CNG (Cryptography Next Generation) APIs where private keys are not exportable.

## Problem

1. Windows machine certificates are stored in `LocalMachine\My` certificate store
2. Private keys are managed by CNG (`ncrypt.dll`) and marked as non-exportable
3. Go's `tls.Certificate` expects either:
   - `PrivateKey` as `crypto.Signer` (preferred)
   - Or raw key bytes (not possible with non-exportable keys)

## Proposed Solution

Implement a `CNG crypto.Signer` wrapper that:
1. Opens the certificate from Windows Cert Store via `crypt32.dll`
2. Gets the private key handle via `ncrypt.dll`
3. Implements `crypto.Signer.Sign()` by calling `NCryptSignHash()`

### Original Interface Sketch

The original sketch below is retained for context. The shipped implementation
uses `WinCertSigner` in `client/internal/auth/wincert_signer.go`, with Windows
certificate discovery in `client/internal/auth/cert_discovery_windows.go`.

```go
// cng_signer_windows.go

package auth

import (
    "crypto"
    "io"
)

// CNGSigner implements crypto.Signer using Windows CNG APIs.
// This allows using non-exportable machine certificates for mTLS.
type CNGSigner struct {
    keyHandle    uintptr  // NCRYPT_KEY_HANDLE
    publicKey    crypto.PublicKey
    certThumbprint string
}

// Public returns the public key.
func (s *CNGSigner) Public() crypto.PublicKey {
    return s.publicKey
}

// Sign signs digest with the private key via NCryptSignHash.
func (s *CNGSigner) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
    // TODO: Implement via NCryptSignHash
    // - Determine padding based on opts (PKCS1v15 vs PSS)
    // - Call NCryptSignHash with appropriate flags
    // - Return signature bytes
}

// NewCNGSignerFromThumbprint loads a certificate by thumbprint and returns a signer.
func NewCNGSignerFromThumbprint(thumbprint string) (*CNGSigner, *x509.Certificate, error) {
    // TODO: Implement
    // 1. CertOpenStore(CERT_STORE_PROV_SYSTEM, "MY", CERT_SYSTEM_STORE_LOCAL_MACHINE)
    // 2. CertFindCertificateInStore(thumbprint)
    // 3. CryptAcquireCertificatePrivateKey()
    // 4. Extract public key from certificate
    // 5. Return CNGSigner wrapping the key handle
}
```

### Required Windows APIs

| API | DLL | Purpose |
|-----|-----|---------|
| `CertOpenStore` | crypt32.dll | Open certificate store |
| `CertFindCertificateInStore` | crypt32.dll | Find cert by thumbprint |
| `CryptAcquireCertificatePrivateKey` | crypt32.dll | Get private key handle |
| `NCryptSignHash` | ncrypt.dll | Sign with CNG key |
| `NCryptFreeObject` | ncrypt.dll | Release key handle |

### Dependencies

```go
import "golang.org/x/sys/windows"
```

## Implementation Notes

### Build Constraints
```go
//go:build windows

package auth
```

### Stub for Non-Windows
```go
//go:build !windows

package auth

func NewCNGSignerFromThumbprint(thumbprint string) (*CNGSigner, *x509.Certificate, error) {
    return nil, nil, errors.New("CNG signer only available on Windows")
}
```

### Testing Requirements
- Requires Windows VM with:
  - AD CS enrolled machine certificate
  - Certificate in `LocalMachine\My` store
  - Non-exportable private key

## Status

**Implemented:** Windows certificate-store discovery and CNG signing are present
in the client auth package.

### Implementation Evidence

- [x] `client/internal/auth/wincert_signer.go` implements `WinCertSigner`,
  `crypto.Signer.Public`, `crypto.Signer.Sign`, RSA/ECDSA signing through
  `NCryptSignHash`, and `Close` cleanup via `NCryptFreeObject`.
- [x] `client/internal/auth/cert_discovery_windows.go` opens
  `LocalMachine\My`, enumerates certificates, supports thumbprint lookup,
  calls `CryptAcquireCertificatePrivateKey`, and returns a `WinCertSigner`.
- [x] `client/internal/auth/wincert_signer_other.go` and
  `client/internal/auth/cert_discovery_other.go` provide non-Windows stubs.
- [x] `client/internal/auth/wincert_pss.go` and
  `client/internal/auth/wincert_pss_test.go` cover RSA-PSS salt handling used
  by NCrypt TLS signing.

### Remaining Validation Notes

1. Full certificate-store behavior still requires a Windows VM with an AD CS
   enrolled machine certificate in `LocalMachine\My`.
2. Cross-platform compile checks should include non-Windows stubs so Linux/macOS
   builds do not import Windows-only APIs.

## References

- [NCryptSignHash](https://learn.microsoft.com/en-us/windows/win32/api/ncrypt/nf-ncrypt-ncryptsignhash)
- [CryptAcquireCertificatePrivateKey](https://learn.microsoft.com/en-us/windows/win32/api/wincrypt/nf-wincrypt-cryptacquirecertificateprivatekey)
- [golang.org/x/sys/windows](https://pkg.go.dev/golang.org/x/sys/windows)
- [Go crypto.Signer](https://pkg.go.dev/crypto#Signer)
