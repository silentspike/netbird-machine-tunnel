# Security Hardening Guide

This guide explains the security features of the NetBird Machine Tunnel and how to configure them properly for production deployments.

All information in this guide has been verified against the actual codebase and tested deployments.

---

## Table of Contents

1. [Security Overview](#security-overview)
2. [Client Security (Windows)](#client-security-windows)
3. [Server Security (Management)](#server-security-management)
4. [Certificate Requirements](#certificate-requirements)
5. [Multi-Tenant Isolation](#multi-tenant-isolation)
6. [Security Checklist](#security-checklist)

---

## Security Overview

The Machine Tunnel uses multiple security layers to protect your infrastructure:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         SECURITY LAYERS                                  │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  CLIENT (Windows)                 SERVER (Management)                    │
│  ─────────────────                ───────────────────                    │
│  • DPAPI Encryption               • mTLS Certificate Validation          │
│  • Machine-bound keys             • Domain-to-Account Mapping            │
│  • Protected config files         • Issuer CA Fingerprint Check          │
│  • WireGuard encryption           • Multi-Tenant Isolation               │
│                                                                          │
│                    ◄─── mTLS (Port 33074) ───►                           │
│                    ◄─── WireGuard Tunnel ───►                            │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

**Key Security Principles:**

| Principle | Implementation |
|-----------|----------------|
| **Encryption at Rest** | DPAPI encrypts sensitive data in config files |
| **Encryption in Transit** | WireGuard tunnel + mTLS for authentication |
| **Machine Binding** | Keys can only be decrypted on the same machine |
| **Certificate-Based Auth** | No passwords, machine certificates from AD CS |
| **Multi-Tenant Isolation** | Domains are mapped to specific accounts |

---

## Client Security (Windows)

### DPAPI Encryption

The Machine Tunnel uses Windows Data Protection API (DPAPI) to encrypt sensitive data. This is a built-in Windows feature that ties encryption to the specific machine.

**What is encrypted:**

| Field | Purpose |
|-------|---------|
| `encrypted_setup_key` | Bootstrap key for initial setup |
| `encrypted_private_key` | WireGuard private key |
| `encrypted_ssh_key` | SSH key for peer connections |

**How it works:**

1. DPAPI uses a machine-specific master key stored by Windows
2. Data encrypted on one machine cannot be decrypted on another
3. Only SYSTEM and Administrators can decrypt the data
4. Additional entropy is compiled into the binary for extra protection

**Verify encryption is working:**

```powershell
# View the config file - keys should be encrypted (Base64 blob, not plaintext)
Get-Content "C:\ProgramData\NetBird\machine-config.yaml"
```

You should see entries like:
```yaml
encrypted_private_key: AQAAANCMnd8BFdERjHoAwE/Cl+sBAAAA...
```

If you see plaintext keys, the encryption is not working correctly.

### File Permissions

The configuration directory should be protected so only SYSTEM and Administrators can access it.

**Verify permissions:**

```powershell
# Check owner and permissions
Get-Acl "C:\ProgramData\NetBird" | Format-List Owner
(Get-Acl "C:\ProgramData\NetBird").Access | Format-Table IdentityReference, FileSystemRights, AccessControlType
```

**Expected output:**

```
Owner: NT AUTHORITY\SYSTEM

IdentityReference       FileSystemRights  AccessControlType
-----------------       ----------------  -----------------
NT AUTHORITY\SYSTEM     FullControl       Allow
BUILTIN\Administrators  FullControl       Allow
```

**If permissions are too open**, restrict them:

```powershell
# Run as Administrator
$path = "C:\ProgramData\NetBird"
$acl = Get-Acl $path

# Remove inheritance
$acl.SetAccessRuleProtection($true, $false)

# Clear existing rules and add only SYSTEM and Administrators
$acl.Access | ForEach-Object { $acl.RemoveAccessRule($_) }

$systemRule = New-Object System.Security.AccessControl.FileSystemAccessRule(
    "NT AUTHORITY\SYSTEM", "FullControl", "ContainerInherit,ObjectInherit", "None", "Allow")
$adminRule = New-Object System.Security.AccessControl.FileSystemAccessRule(
    "BUILTIN\Administrators", "FullControl", "ContainerInherit,ObjectInherit", "None", "Allow")

$acl.SetAccessRule($systemRule)
$acl.SetAccessRule($adminRule)
Set-Acl -Path $path -AclObject $acl
```

### Service Account

The Machine Tunnel service runs as `NT AUTHORITY\SYSTEM`. This is required because:

1. The service needs to start before any user logs in
2. DPAPI machine-scope encryption requires SYSTEM context
3. The service needs access to the machine certificate store

**Verify service account:**

```powershell
Get-WmiObject Win32_Service -Filter "Name='NetBirdMachine'" | Select-Object StartName
```

Expected: `LocalSystem`

---

## Server Security (Management)

### mTLS Authentication

The management server uses mutual TLS (mTLS) on a dedicated port for machine authentication. This means:

1. The server presents its certificate to the client
2. The client presents its machine certificate to the server
3. Both sides verify the other's certificate

**Dedicated mTLS Port:**

| Port | Purpose | Authentication |
|------|---------|----------------|
| 443 | Standard NetBird clients | Setup keys, SSO tokens |
| 33074 | Machine Tunnel only | Client certificates required |

**Verify mTLS is running:**

```bash
# On the management server
ss -tlnp | grep 33074

# Expected output:
# LISTEN  0  4096  0.0.0.0:33074  0.0.0.0:*
```

**Check server logs for mTLS activity:**

```bash
docker logs netbird-management --tail 100 | grep -i mtls
```

You should see:
```
INFO server/mtls_server.go: mTLS server configured: port=33074, CA pool loaded with 1 certificates
INFO server/mtls_server.go: starting mTLS-only Machine Tunnel server on port 33074
```

### CA Trust Configuration

The server must be configured to trust your organization's Certificate Authority (CA).

**Server configuration (environment variables or config file):**

| Setting | Description |
|---------|-------------|
| `NETBIRD_MTLS_CA_FILE` | Path to CA certificate file |
| `NETBIRD_MTLS_ENABLED` | Enable mTLS (true/false) |

**Verify CA is loaded:**

```bash
docker logs netbird-management | grep "CA certificate"
```

You should see:
```
INFO server/mtls_server.go: loaded CA certificate: your-ca.crt
INFO server/mtls_server.go: mTLS CA pool loaded: 1 certificates
```

---

## Certificate Requirements

### Machine Certificate

Each Windows machine needs a certificate for authentication. The certificate must meet these requirements:

**Required Certificate Properties:**

| Property | Requirement | Example |
|----------|-------------|---------|
| **Subject** | Contains computer name | `CN=DESKTOP-ABC123.corp.local` |
| **Subject Alternative Name (SAN)** | DNS name = hostname.domain | `DNS Name=DESKTOP-ABC123.corp.local` |
| **Enhanced Key Usage** | Client Authentication | OID: 1.3.6.1.5.5.7.3.2 |
| **Private Key** | Must exist, should be non-exportable | `HasPrivateKey: True` |

**Verify your certificate:**

```powershell
# Find the machine certificate
$cert = Get-ChildItem Cert:\LocalMachine\My | Where-Object { $_.Subject -match $env:COMPUTERNAME }

# Check basic properties
$cert | Format-List Subject, Issuer, NotBefore, NotAfter, HasPrivateKey

# Check SAN (Subject Alternative Name) - THIS IS CRITICAL
$cert.Extensions | Where-Object { $_.Oid.FriendlyName -eq "Subject Alternative Name" } |
    ForEach-Object { $_.Format($true) }
```

**Expected output:**

```
Subject       : CN=DESKTOP-ABC123.corp.local
Issuer        : CN=Your-CA, DC=corp, DC=local
NotBefore     : 1/26/2026 11:46:13 PM
NotAfter      : 1/26/2027 11:46:13 PM
HasPrivateKey : True

DNS Name=DESKTOP-ABC123.corp.local
```

**Common Certificate Problems:**

| Problem | Symptom | Solution |
|---------|---------|----------|
| Missing SAN | Server rejects certificate | Update certificate template to include SAN |
| Wrong SAN format | Authentication fails | SAN must be `hostname.domain`, not just hostname |
| Expired certificate | Connection refused | Renew certificate |
| No private key | Cannot authenticate | Re-enroll certificate |
| Wrong issuer CA | Server doesn't trust it | Add CA to server's trust store |

### AD CS Template Configuration

If you're using Active Directory Certificate Services (AD CS), configure your template with:

1. **Template name:** Create a new template (e.g., "NetBird Machine")
2. **Subject name:** Build from Active Directory information
3. **Subject Alternative Name:** Include DNS name from AD
4. **Security:** Allow "Domain Computers" to enroll
5. **Private key:** Mark as non-exportable for security

---

## Multi-Tenant Isolation

The Machine Tunnel supports multiple tenants (organizations) on the same management server. Each tenant is isolated by:

1. **Domain-to-Account Mapping:** Each AD domain is mapped to a specific NetBird account
2. **Allowed Domains:** Each account has a list of allowed domains
3. **Issuer Validation:** Each account can require specific CA fingerprints

### How It Works

```
Certificate SAN: workstation1.company-a.local
                        │
                        ▼
              ┌─────────────────────┐
              │ Domain Extraction   │
              │ company-a.local     │
              └─────────────────────┘
                        │
                        ▼
              ┌─────────────────────┐
              │ Domain → Account    │
              │ Mapping Lookup      │
              └─────────────────────┘
                        │
                        ▼
              ┌─────────────────────┐
              │ Is domain in        │
              │ AllowedDomains?     │──No──► REJECT
              └─────────────────────┘
                        │ Yes
                        ▼
              ┌─────────────────────┐
              │ Is issuer CA in     │
              │ AllowedIssuers?     │──No──► REJECT
              └─────────────────────┘
                        │ Yes
                        ▼
                    ALLOW
```

### Security Boundaries

| Check | Purpose | What Happens If Failed |
|-------|---------|------------------------|
| Domain Mapping | Route to correct account | "domain not mapped to any account" |
| Allowed Domains | Prevent cross-tenant access | "domain not in allowed list" |
| Issuer Fingerprint | Prevent rogue CA attacks | "issuer CA not in allowed list" |

**View authentication in server logs:**

```bash
docker logs netbird-management --tail 100 | grep -E "peer registration|domain|account"
```

Successful authentication looks like:
```
INFO grpc/machine_tunnel.go: Machine peer registration: hostname=DESKTOP-ABC domain=corp.local account=abc123...
```

---

## Security Checklist

Use this checklist when deploying the Machine Tunnel in production.

### Client (Windows)

- [ ] **Config file encrypted** - Check that `encrypted_private_key` contains DPAPI blob, not plaintext
- [ ] **Permissions restricted** - Only SYSTEM and Administrators should have access to `C:\ProgramData\NetBird`
- [ ] **Service runs as SYSTEM** - Required for DPAPI and pre-login operation
- [ ] **Certificate has correct SAN** - Must contain `hostname.domain` format
- [ ] **Certificate not expired** - Check validity dates
- [ ] **Private key non-exportable** - Recommended for production (check certificate template)

### Server (Management)

- [ ] **mTLS port listening** - Port 33074 should be open and listening
- [ ] **CA certificate loaded** - Check logs for "CA pool loaded"
- [ ] **Domain mapping configured** - Each client domain must map to an account
- [ ] **Allowed domains set** - Each account should have explicit allowed domains
- [ ] **Issuer fingerprints set** - Recommended for production (prevents rogue CA)
- [ ] **TLS certificate valid** - Management server should use proper TLS cert (not self-signed in production)

### Network

- [ ] **Port 33074 accessible** - Clients must reach management server on this port
- [ ] **Firewall allows mTLS** - Corporate firewalls may need configuration
- [ ] **No TLS inspection** - mTLS will fail if traffic is inspected/modified

### Verification Commands

**Client:**
```powershell
# All-in-one client security check
Write-Host "=== NetBird Machine Tunnel Security Check ===" -ForegroundColor Cyan

# 1. Service account
$svc = Get-WmiObject Win32_Service -Filter "Name='NetBirdMachine'"
Write-Host "`nService Account: $($svc.StartName)" -ForegroundColor $(if($svc.StartName -eq 'LocalSystem'){'Green'}else{'Red'})

# 2. Config encryption
$config = Get-Content "C:\ProgramData\NetBird\machine-config.yaml" -Raw
$encrypted = $config -match 'encrypted_private_key: AQAAANC'
Write-Host "Config Encrypted: $encrypted" -ForegroundColor $(if($encrypted){'Green'}else{'Red'})

# 3. File permissions
$owner = (Get-Acl "C:\ProgramData\NetBird").Owner
Write-Host "Config Owner: $owner" -ForegroundColor $(if($owner -eq 'NT AUTHORITY\SYSTEM'){'Green'}else{'Yellow'})

# 4. Certificate
$cert = Get-ChildItem Cert:\LocalMachine\My | Where-Object { $_.Subject -match $env:COMPUTERNAME }
if ($cert) {
    $san = $cert.Extensions | Where-Object { $_.Oid.FriendlyName -eq "Subject Alternative Name" }
    Write-Host "Certificate Found: $($cert.Subject)" -ForegroundColor Green
    Write-Host "Certificate Valid Until: $($cert.NotAfter)" -ForegroundColor $(if($cert.NotAfter -gt (Get-Date)){'Green'}else{'Red'})
    Write-Host "SAN Present: $($san -ne $null)" -ForegroundColor $(if($san){'Green'}else{'Red'})
} else {
    Write-Host "Certificate: NOT FOUND" -ForegroundColor Red
}
```

**Server:**
```bash
# All-in-one server security check
echo "=== NetBird Management Server Security Check ==="

# 1. mTLS port
echo -n "mTLS Port 33074: "
ss -tlnp | grep -q 33074 && echo "LISTENING" || echo "NOT LISTENING"

# 2. CA loaded
echo -n "CA Certificate: "
docker logs netbird-management 2>&1 | grep -q "CA pool loaded" && echo "LOADED" || echo "NOT LOADED"

# 3. mTLS config
echo -n "mTLS Config: "
docker logs netbird-management 2>&1 | grep -q "mTLS config loaded" && echo "CONFIGURED" || echo "NOT CONFIGURED"
```

---

## Related Documentation

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture and design decisions
- [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Diagnosing and fixing problems
- [NOTICE.md](../NOTICE.md) - License information
