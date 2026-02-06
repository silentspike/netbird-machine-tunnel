---
name: Machine Tunnel Bug Report
about: Report an issue specific to the Machine Tunnel feature (pre-login VPN)
title: '[Bug] '
labels: type:bug, component:client
assignees: ''
---

## Description

A clear and concise description of what the bug is.

## Component

Which Machine Tunnel component is affected?

- [ ] Windows Service (`NetBirdMachine`)
- [ ] mTLS Authentication / Certificate Enrollment
- [ ] WireGuard Tunnel (`wg-nb-machine` interface)
- [ ] NRPT / DNS Configuration
- [ ] Firewall Rules
- [ ] Domain Join / Kerberos
- [ ] Bootstrap (Setup Key phase)
- [ ] Reconnection / Health Checks
- [ ] Other: ___

## Steps to Reproduce

1. ...
2. ...
3. ...

## Expected Behavior

What should happen?

## Actual Behavior

What happens instead?

## Environment

- **Windows Version:** [e.g., Windows 11 23H2, Server 2022]
- **NetBird Machine Tunnel Version:** [e.g., 1.0.0]
- **Domain:** [e.g., Active Directory domain-joined, standalone]
- **Certificate Type:** [e.g., AD CS auto-enrolled, manual, setup key only]
- **Network:** [e.g., corporate LAN, remote/VPN, split-tunnel]

## Diagnostic Output

Please run the following commands and attach the output:

```powershell
# Service status
Get-Service NetBirdMachine | Format-List *

# Interface status
Get-NetAdapter -Name "wg-nb-machine" -ErrorAction SilentlyContinue | Format-List Status, InterfaceGuid

# Recent event log entries (last 20)
Get-WinEvent -FilterHashtable @{LogName='Application'; ProviderName='NetBirdMachine'} -MaxEvents 20 -ErrorAction SilentlyContinue | Format-Table TimeCreated, LevelDisplayName, Message -Wrap

# NRPT rules
Get-DnsClientNrptRule | Where-Object { $_.Comment -like "NetBird*" }

# Firewall rules
Get-NetFirewallRule -Group "NetBird Machine Tunnel" -ErrorAction SilentlyContinue | Format-Table DisplayName, Action, Direction
```

## Logs

<details>
<summary>Machine Tunnel log (C:\ProgramData\NetBird\machine-tunnel.log)</summary>

```
Paste relevant log lines here. Redact any private keys, certificates, or internal hostnames.
```

</details>

## Additional Context

Add any other context, screenshots, or network diagrams here.

## Troubleshooting Checklist

- [ ] Reviewed the [Troubleshooting Guide](../../docs/TROUBLESHOOTING.md)
- [ ] Checked Windows Event Log for `NetBirdMachine` entries
- [ ] Verified certificate is valid and not expired (`certutil -store My`)
- [ ] Confirmed DC connectivity when tunnel is up (`Test-NetConnection dc01 -Port 389`)
- [ ] Searched for similar issues on GitHub (including closed ones)
