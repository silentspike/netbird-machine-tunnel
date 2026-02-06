# Lab Configuration Template
# Copy this file to lab-config.ps1 and fill in your actual values.
# lab-config.ps1 is in .gitignore and will NOT be committed.

# Management Server
$LabManagementServer = "mgmt.example.com"
$LabManagementPort = 443

# Domain Controller
$LabDCAddress = "dc.example.com"
$LabDomainName = "corp.example.com"

# Proxmox Host (for E2E test automation)
$LabProxmoxHost = "proxmox.example.com"
$LabProxmoxUser = "root"

# Windows Test VM
$LabWinVMID = 102
$LabWinVMIP = "win-client.example.com"
$LabWinVMUser = "admin"

# Router Peer
$LabRouterPeerIP = "router.example.com"

# NetBird Setup Key (generate fresh ones, 24h TTL!)
# $LabSetupKey = ""
