# Notice

This project is a fork of [NetBird](https://github.com/netbirdio/netbird), licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).

## Original Project

- **Project:** NetBird
- **Repository:** https://github.com/netbirdio/netbird
- **License:** AGPL-3.0
- **Copyright:** NetBird Authors

## Fork Information

- **Fork Repository:** https://github.com/obtFusi/netbird-fork
- **Fork Purpose:** Machine Tunnel for Windows Pre-Login VPN with AD Certificate Authentication

## Modifications

This fork includes the following modifications from upstream NetBird:

### Machine Tunnel Feature (client/internal/tunnel/)
- **mTLS Authentication:** Machine certificate authentication using Windows AD CS
- **NRPT Integration:** DNS Name Resolution Policy Table for AD domain routing
- **Windows Firewall Rules:** Granular port-based ACLs for Domain Controller access
- **Health Monitoring:** Tunnel health checks with auto-reconnect
- **Windows Event Log:** Service events logged to Windows Event Viewer
- **Crash Recovery:** Automatic reconnection with exponential backoff
- **DPAPI Security:** Secure key storage using Windows Data Protection API

### Windows Service (client/cmd/netbird-machine/)
- **SYSTEM Service:** Runs as NT AUTHORITY\SYSTEM for pre-login operation
- **SCM Integration:** Windows Service Control Manager with recovery actions
- **Bootstrap Flow:** Two-phase authentication (Setup Key → mTLS)

### Server Extensions (management/server/)
- **Machine Peer Registration:** RegisterMachinePeer RPC for machine authentication
- **mTLS Validation:** Certificate validation with AllowedDomains per account
- **Issuer Fingerprint:** Strong binding via certificate chain validation

## Source Code Availability

As required by AGPL-3.0, the complete source code for this fork is available at:

https://github.com/obtFusi/netbird-fork

## Original Authors

NetBird was created by the NetBird team and contributors:
https://github.com/netbirdio/netbird/graphs/contributors

## License

This fork is distributed under the same AGPL-3.0 license as the original NetBird project. See [LICENSE](LICENSE) for the full license text.
