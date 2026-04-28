# ADR-001: mTLS Port Strategy

**Status:** Superseded by dedicated mTLS port implementation
**Date:** 2026-01-19
**Issue:** #14 (T-1.2: NetBird gRPC mTLS Interception)

> 2026-04-27 update: this ADR records the original decision, but the shipped
> implementation uses a dedicated mTLS-only gRPC server on port `33074`.
> See `management/internals/server/mtls_server.go` and
> `management/internals/server/server.go`.

## Context

The Machine Tunnel Fork adds mTLS (mutual TLS) authentication for machine peers using AD CS certificates. We need to decide how to expose the mTLS-authenticated gRPC endpoints.

**Options considered:**

### Option A: Single Port with Method-Based Routing
- Standard gRPC port (33073) handles both token-auth and mTLS
- TLS config uses `VerifyClientCertIfGiven` (not `RequireAndVerifyClientCert`)
- gRPC interceptors check if method requires mTLS
- Methods in `mTLSRequiredMethods` map reject requests without valid cert
- Other methods fall back to token authentication

### Option B: Dual Port
- Standard port (33073) for token authentication
- Separate port (33074) for mTLS-only authentication
- Each port has its own TLS config
- Simpler routing but more network config

## Decision

**Historical decision:** Option A, single port with method-based routing.

**Current implementation:** Option B, a dedicated mTLS-only port. The dedicated
server defaults to `33074`, uses `tls.RequireAndVerifyClientCert`, and is
started alongside the standard management server. The method map in
`management/internals/server/mtls_auth.go` still defines the Machine Tunnel RPCs
that require mTLS identity, but transport separation is now enforced by the
dedicated server.

## Historical Rationale

1. **Simpler Deployment:** Only one port to configure in firewalls and load balancers
2. **Graceful Fallback:** Machines without certificates can still use Setup-Key auth for bootstrap
3. **Existing Infrastructure:** Works with existing NetBird deployments without port changes
4. **TLS Flexibility:** `VerifyClientCertIfGiven` allows:
   - Clients WITH certs: Verified against CA pool, mTLS identity extracted
   - Clients WITHOUT certs: TLS handshake succeeds, fall back to token auth

The rationale above explains why Option A was selected originally. It is not
the current enforcement model.

## Current Rationale

The shipped implementation uses Option B because Machine Tunnel RPCs are easier
to reason about when they are exposed on a dedicated mTLS-only listener:

1. **Strict transport boundary:** Port `33074` uses
   `tls.RequireAndVerifyClientCert`, so Machine Tunnel clients cannot reach the
   service without a verified client certificate.
2. **Standard management compatibility:** The regular management listener keeps
   the normal NetBird token-auth flow.
3. **Clear operations model:** Firewalls and load balancers can route machine
   certificate traffic separately from regular management traffic.
4. **Explicit bootstrap split:** Setup-key bootstrap remains separate from the
   certificate-backed Machine Tunnel RPC path.

## Historical Implementation Sketch

The following sketch belongs to the original single-port decision. The current
implementation evidence below is authoritative for the shipped design.

### TLS Configuration (boot.go)
```go
config := &tls.Config{
    ClientAuth: tls.VerifyClientCertIfGiven,  // NOT RequireAndVerifyClientCert
    ClientCAs:  caCertPool,
    // ...
}
```

### Method-Based Requirements (mtls_auth.go)
```go
var mTLSRequiredMethods = map[string]bool{
    "/management.ManagementService/RegisterMachinePeer": true,
    "/management.ManagementService/SyncMachinePeer":     true,
    "/management.ManagementService/GetMachineRoutes":    true,
    "/management.ManagementService/ReportMachineStatus": true,
}
```

### Interceptor Logic
```
Request arrives
  |
  +-> Extract mTLS identity from TLS state
  |     |
  |     +-> Success: Identity available
  |     |     |
  |     |     +-> Store in context, proceed
  |     |
  |     +-> Failure: No cert or invalid
  |           |
  |           +-> Method requires mTLS?
  |           |     |
  |           |     +-> YES: Return Unauthenticated error
  |           |     +-> NO:  Fall back to token auth
```

## Consequences

### Positive
- Dedicated mTLS listener uses `RequireAndVerifyClientCert`.
- Standard management traffic and machine-certificate traffic are separated.
- Existing NetBird token-auth flows remain on the normal management listener.
- `mTLSRequiredMethods` remains useful as a method-level guard and interceptor
  source of truth, even though transport separation is now the first boundary.

### Negative
- Deployments need to expose and protect an additional management-side port.
- Operators must configure mTLS CA material and domain/account mappings before
  enabling the Machine Tunnel listener.

### Risks
- **Misconfiguration:** If `mTLSRequiredMethods` is not kept in sync with proto definitions
  - Mitigation: Code review, integration tests
- **CA Pool exhaustion:** Large multi-tenant deployments with many CAs
  - Mitigation: MTLSCADir supports multiple CA files, can scale
- **Port exposure:** If `33074` is exposed without the intended CA and
  account-domain configuration, machine authentication will fail closed but
  operators may see unavailable Machine Tunnel RPCs
  - Mitigation: pre-flight config checks and lab smoke before go-live

## Related Decisions

- ADR-002 (implemented): Windows CNG crypto.Signer Interface
- Potential future ADR: Multi-Tenant CA Isolation, if the implementation needs
  a separate architectural decision beyond the current per-account issuer and
  domain controls.

## Implementation Status

### Historical Implementation Artifacts

These entries belong to the original single-port design and are retained for
history. They are not the authoritative current transport boundary.

- [x] `management/internals/server/mtls_auth.go` - gRPC interceptors
- [x] `management/internals/server/config/config.go` - mTLS config fields
- [x] `management/internals/server/boot.go` - legacy mTLS config and
  interceptor chain for the standard management server
- [x] `shared/management/proto/management.proto` - Machine RPC definitions

### Current Implementation Evidence

- [x] `management/internals/server/mtls_server.go` - dedicated mTLS-only server,
  default port `33074`, `tls.RequireAndVerifyClientCert`.
- [x] `management/internals/server/server.go` - starts and stops the separate
  mTLS server with the main management server lifecycle.
- [x] `management/internals/server/mtls_auth.go` - Machine Tunnel RPC method map
  and unary/stream interceptors.
- [x] `shared/management/proto/management.proto` - Machine Tunnel RPC
  definitions.
- [x] `shared/management/proto/management.pb.go` and
  `shared/management/proto/management_grpc.pb.go` - generated code for
  `RegisterMachinePeer`, `SyncMachinePeer`, `GetMachineRoutes`, and
  `ReportMachineStatus`.

## References

- [Go TLS ClientAuthType](https://pkg.go.dev/crypto/tls#ClientAuthType)
- [gRPC Interceptors](https://grpc.io/docs/guides/interceptors/)
- [Dedicated mTLS server](../management/internals/server/mtls_server.go)
- [Management server lifecycle](../management/internals/server/server.go)
- [Legacy NetBird gRPC server setup](../management/internals/server/boot.go)
