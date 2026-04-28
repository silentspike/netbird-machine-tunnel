# AD CS Machine Certificate Template

## Use Case

Create an Active Directory Certificate Services template that issues
non-exportable machine certificates for NetBird Machine Tunnel mTLS
authentication.

## Preconditions

- Active Directory Certificate Services is already deployed.
- Windows machines are domain joined before certificate enrollment.
- The issuing CA certificate is configured on the NetBird management server as
  the Machine Tunnel mTLS trust anchor.
- The certificate subject or SAN can produce the machine DNS name used by your
  NetBird account's allowed domain mapping.

## Minimal Template Settings

Use the Microsoft Certificate Templates console (`certtmpl.msc`):

1. Duplicate the built-in `Computer` template.
2. Name the duplicate template clearly, for example `NetBird Machine Tunnel`.
3. On `Compatibility`, choose values supported by your domain controllers and
   enrolled Windows clients.
4. On `Subject Name`, select `Build from this Active Directory information`.
5. Include DNS name in the Subject Alternative Name.
6. On `Extensions`, ensure `Client Authentication` is present in Enhanced Key
   Usage.
7. On `Request Handling`, keep private keys non-exportable.
8. On `Cryptography`, use a provider and key size accepted by your Windows
   fleet and CA policy.
9. On `Security`, grant `Enroll` and, if using auto-enrollment, `Autoenroll` to
   the computer group that should receive Machine Tunnel certificates.
10. Publish the template on the issuing CA.

## Expected Verification

On an enrolled Windows machine:

```powershell
Get-ChildItem Cert:\LocalMachine\My |
  Where-Object { $_.EnhancedKeyUsageList.FriendlyName -contains "Client Authentication" } |
  Select-Object Subject, DnsNameList, NotAfter, Thumbprint
```

Expected result:

- The certificate is in `LocalMachine\My`.
- `DnsNameList` includes the machine DNS name expected by the NetBird account.
- The certificate includes Client Authentication EKU.
- The private key is present but not exportable.

## Security Caveats

- Do not issue Machine Tunnel certificates to broad user groups.
- Prefer a dedicated computer group for Machine Tunnel enrollment.
- Keep certificate validity short enough for operational rotation.
- Revocation checking is not implemented by the Machine Tunnel mTLS path yet;
  compensate with short validity, issuer fingerprint constraints, and per-account
  allowed-domain scoping.
- Do not export private keys into deployment shares, scripts, or tickets.
