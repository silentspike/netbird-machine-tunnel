// Machine Tunnel Fork - Windows Firewall Manager
// Manages Windows Firewall rules for DC traffic over the Machine Tunnel.
//
// Enterprise SOTA: Uses PowerShell New-NetFirewallRule with -InterfaceAlias
// to scope rules to the WireGuard interface only. This prevents blocking
// system-wide traffic (e.g., management server connectivity).
//
// References:
// - T-4.5: Firewall Allow Rules for DC traffic
// - T-4.6: Deny-Default rule (interface-scoped)

//go:build windows

package tunnel

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	// FirewallRulePrefix is the prefix for Machine Tunnel firewall rules.
	FirewallRulePrefix = "NetBird-Machine-"

	// FirewallGroupName is the group name for all Machine Tunnel rules.
	// PowerShell New-NetFirewallRule supports -Group for easy management.
	FirewallGroupName = "NetBird Machine Tunnel"
)

// FirewallManager handles Windows Firewall rules for DC traffic
type FirewallManager struct {
	interfaceName string
	rulePrefix    string
	groupName     string
}

// NewFirewallManager creates a new Firewall manager
func NewFirewallManager(interfaceName string) *FirewallManager {
	return &FirewallManager{
		interfaceName: interfaceName,
		rulePrefix:    FirewallRulePrefix,
		groupName:     FirewallGroupName,
	}
}

// AllowDCTraffic adds firewall rules to allow traffic to a DC IP/CIDR.
// Rules are scoped to the WireGuard interface via -InterfaceAlias.
func (m *FirewallManager) AllowDCTraffic(ipOrCIDR string) error {
	ruleName := m.generateRuleName(ipOrCIDR, "allow")

	log.WithFields(log.Fields{
		"ip_or_cidr": ipOrCIDR,
		"rule_name":  ruleName,
		"interface":  m.interfaceName,
	}).Info("Adding firewall allow rule for DC traffic")

	// Delete existing rules first (ignore errors)
	_ = m.deleteRuleByName(ruleName + "-in")
	_ = m.deleteRuleByName(ruleName + "-out")

	// Add inbound allow rule - scoped to WireGuard interface
	if err := m.addPowerShellRule(ruleName+"-in", "Inbound", "Allow", ipOrCIDR); err != nil {
		return fmt.Errorf("add inbound rule: %w", err)
	}

	// Add outbound allow rule - scoped to WireGuard interface
	if err := m.addPowerShellRule(ruleName+"-out", "Outbound", "Allow", ipOrCIDR); err != nil {
		return fmt.Errorf("add outbound rule: %w", err)
	}

	log.WithField("rule_name", ruleName).Info("Firewall rules added successfully")
	return nil
}

// EnableDenyDefault enables deny-default rules for the Machine Tunnel interface.
// This implements T-4.6: deny-by-default security for tunnel traffic.
//
// CRITICAL: Rules are scoped to the WireGuard interface (-InterfaceAlias).
// This ensures only tunnel traffic is affected, not system-wide traffic
// (e.g., management server, Signal server connectivity).
func (m *FirewallManager) EnableDenyDefault() error {
	if m.interfaceName == "" {
		log.Warn("No interface name configured, skipping deny-default rules")
		return nil
	}

	ruleName := m.rulePrefix + "deny-default"

	log.WithFields(log.Fields{
		"rule_name": ruleName,
		"interface": m.interfaceName,
	}).Info("Enabling deny-default firewall rule (interface-scoped)")

	// Delete existing rules first
	_ = m.deleteRuleByName(ruleName + "-in")
	_ = m.deleteRuleByName(ruleName + "-out")

	// Add deny-default inbound rule - scoped to WireGuard interface
	// RemoteAddress=Any blocks all inbound traffic on this interface
	// that doesn't match a more specific allow rule
	psScriptIn := fmt.Sprintf(`
		New-NetFirewallRule -DisplayName '%s-in' `+
		`-Name '%s-in' `+
		`-Group '%s' `+
		`-Direction Inbound `+
		`-Action Block `+
		`-InterfaceAlias '%s' `+
		`-RemoteAddress Any `+
		`-Protocol Any `+
		`-ErrorAction Stop`,
		ruleName, ruleName, m.groupName, m.interfaceName)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScriptIn)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PowerShell add deny-default inbound failed: %w, output: %s", err, string(output))
	}

	// Add deny-default outbound rule - scoped to WireGuard interface
	psScriptOut := fmt.Sprintf(`
		New-NetFirewallRule -DisplayName '%s-out' `+
		`-Name '%s-out' `+
		`-Group '%s' `+
		`-Direction Outbound `+
		`-Action Block `+
		`-InterfaceAlias '%s' `+
		`-RemoteAddress Any `+
		`-Protocol Any `+
		`-ErrorAction Stop`,
		ruleName, ruleName, m.groupName, m.interfaceName)

	cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScriptOut)
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PowerShell add deny-default outbound failed: %w, output: %s", err, string(output))
	}

	log.Info("Deny-default firewall rules enabled (interface-scoped)")
	return nil
}

// addPowerShellRule adds a firewall rule using PowerShell New-NetFirewallRule.
// Rules are scoped to the WireGuard interface via -InterfaceAlias.
func (m *FirewallManager) addPowerShellRule(ruleName, direction, action, remoteAddress string) error {
	// PowerShell script to create firewall rule with interface binding
	psScript := fmt.Sprintf(`
		New-NetFirewallRule -DisplayName '%s' `+
		`-Name '%s' `+
		`-Group '%s' `+
		`-Direction %s `+
		`-Action %s `+
		`-InterfaceAlias '%s' `+
		`-RemoteAddress '%s' `+
		`-Protocol Any `+
		`-ErrorAction Stop`,
		ruleName, ruleName, m.groupName, direction, action, m.interfaceName, remoteAddress)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("PowerShell add rule failed: %w, output: %s", err, string(output))
	}

	return nil
}

// deleteRuleByName deletes a firewall rule by its name using PowerShell.
func (m *FirewallManager) deleteRuleByName(ruleName string) error {
	psScript := fmt.Sprintf(`Remove-NetFirewallRule -Name '%s' -ErrorAction SilentlyContinue`, ruleName)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	_, err := cmd.CombinedOutput()
	return err
}

// RemoveAllRules removes all Machine Tunnel firewall rules.
// Uses PowerShell Get-NetFirewallRule with -Group for efficient cleanup.
func (m *FirewallManager) RemoveAllRules() error {
	log.Info("Removing all Machine Tunnel firewall rules")

	// First try to remove by group (most efficient)
	psScript := fmt.Sprintf(`
		$rules = Get-NetFirewallRule -Group '%s' -ErrorAction SilentlyContinue
		if ($rules) {
			$rules | Remove-NetFirewallRule -ErrorAction SilentlyContinue
			Write-Output "Removed $($rules.Count) rules"
		} else {
			Write-Output "No rules found in group"
		}
	`, m.groupName)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"output": string(output),
		}).Warn("Failed to remove rules by group, trying by prefix")
	} else {
		log.WithField("output", strings.TrimSpace(string(output))).Debug("Group-based cleanup result")
	}

	// Also try to remove by name prefix (fallback for rules created before group support)
	psScriptPrefix := fmt.Sprintf(`
		$rules = Get-NetFirewallRule -Name '%s*' -ErrorAction SilentlyContinue
		if ($rules) {
			$rules | Remove-NetFirewallRule -ErrorAction SilentlyContinue
			Write-Output "Removed $($rules.Count) rules by prefix"
		}
	`, m.rulePrefix)

	cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScriptPrefix)
	output, err = cmd.CombinedOutput()
	if err != nil {
		log.WithFields(log.Fields{
			"error":  err,
			"output": string(output),
		}).Warn("Failed to remove rules by prefix")
	}

	log.Info("Machine Tunnel firewall rules cleanup complete")
	return nil
}

// ListRules lists all Machine Tunnel firewall rules.
func (m *FirewallManager) ListRules() ([]string, error) {
	psScript := fmt.Sprintf(`
		$rules = @()
		$rules += Get-NetFirewallRule -Group '%s' -ErrorAction SilentlyContinue
		$rules += Get-NetFirewallRule -Name '%s*' -ErrorAction SilentlyContinue
		$rules | Select-Object -ExpandProperty Name -Unique
	`, m.groupName, m.rulePrefix)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("PowerShell list rules failed: %w, output: %s", err, string(output))
	}

	// Parse output - each line is a rule name
	var rules []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			rules = append(rules, line)
		}
	}

	return rules, nil
}

// generateRuleName generates a unique firewall rule name
func (m *FirewallManager) generateRuleName(ipOrCIDR, action string) string {
	// Sanitize IP/CIDR for use in rule name
	sanitized := strings.ReplaceAll(ipOrCIDR, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	return fmt.Sprintf("%s%s-%s", m.rulePrefix, action, sanitized)
}
