# Group Policy Deployment Example

## Use Case

Deploy the NetBird Machine Tunnel Windows service to domain-joined computers so
the tunnel can start before user login and provide Active Directory
connectivity.

## Preconditions

- `netbird-machine.exe` is already built or downloaded from a trusted release
  artifact.
- The target computers are in a dedicated security group or OU.
- AD CS machine certificate enrollment is configured.
- A short-lived setup key is available only for the bootstrap window.
- Your management URL and mTLS settings are already configured on the server.

## Minimal Steps

1. Create a Group Policy Object scoped to the target computer OU.
2. Copy `netbird-machine.exe` and the startup script to a controlled deployment
   share, for example:

   ```text
   \\<domain-fqdn>\SYSVOL\<domain-fqdn>\scripts\NetBirdMachine\
   ```

3. Add a Computer Startup script that installs the service if it is missing:

   ```powershell
   $InstallRoot = "\\<domain-fqdn>\SYSVOL\<domain-fqdn>\scripts\NetBirdMachine"
   $Binary = Join-Path $InstallRoot "netbird-machine.exe"
   $ConfigDir = "C:\ProgramData\NetBird"
   $ConfigFile = Join-Path $ConfigDir "machine-config.yaml"

   New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null

   if (-not (Test-Path $ConfigFile)) {
     @"
management_url: "https://<management-hostname>"
setup_key: "<short-lived-setup-key>"
machine_cert_enabled: false
"@ | Set-Content -Path $ConfigFile -Encoding UTF8 -NoNewline
   }

   & $Binary install
   Start-Service NetBirdMachine
   ```

4. After machines transition to mTLS, revoke the setup key and confirm the
   config no longer contains plaintext `setup_key`.

## Expected Verification

On a target Windows machine:

```powershell
gpresult /r /scope computer
Get-Service NetBirdMachine
Get-Content "C:\ProgramData\NetBird\machine-config.yaml"
Test-NetConnection -ComputerName <domain-controller-fqdn> -Port 389
```

Expected result:

- The target GPO is applied under Computer Settings.
- `NetBirdMachine` is `Running` and `Automatic`.
- The config uses DPAPI-encrypted values after first successful service start.
- Required Domain Controller ports are reachable through the Machine Tunnel.

## Security Caveats

- Do not place long-lived setup keys in SYSVOL.
- Prefer a short-lived setup key with a tight usage limit for each deployment
  wave.
- Revoke the setup key after the mTLS transition.
- Restrict the deployment share to administrators and target computers.
- Do not deploy customer hostnames, real setup keys, or private CA material in
  public examples or tickets.
