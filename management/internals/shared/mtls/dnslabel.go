package mtls

// Machine Tunnel Fork - DNSLabel Generation for mTLS Peers
// Provides unique DNS label generation to prevent collisions across domains.

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// dnsLabelRegex validates RFC 1123 compliant DNS labels
// Must start and end with alphanumeric, can contain hyphens in between
var dnsLabelRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// MaxDNSLabelLength is the maximum length for a DNS label per RFC 1123
const MaxDNSLabelLength = 63

// HashSuffixLength is the length of the FQDN hash suffix (16 hex chars = 64 bits)
const HashSuffixLength = 16

const (
	dnsLabelHMACKey = "netbird-machine-tunnel-dns-label-v1"
	hashSuffixBytes = HashSuffixLength / 2
)

// GenerateUniqueDNSLabel creates a unique DNSLabel from hostname and domain.
//
// Problem (v3.5 - domain-only hash):
//   - Hash only over domain → all hosts of a domain get same hash suffix
//   - Hostname collision within domain would result in identical DNSLabel
//   - Example: Two "win10-pc.corp.local" (misconfiguration) → same DNSLabel!
//
// Solution (v3.6 - FQDN hash):
//   - FQDN = "hostname.domain" (case-insensitive)
//   - Each host gets guaranteed unique hash
//   - "win10-pc.customer-a.local" → "win10-pc-a1b2c3d4e5f60718"
//   - "win10-pc.customer-b.local" → "win10-pc-5e6f708192a3b4c5"
//   - "win11-pc.customer-a.local" → "win11-pc-9a8b7c6d5e4f3021"
//
// The suffix is a deterministic HMAC-based fingerprint, not an authentication
// secret. The public HMAC key is only a domain-separation label for this use.
// Hash collision probability with 64 bits (16 hex chars) and 10,000 peers is
// negligible for DNS label generation.
func GenerateUniqueDNSLabel(hostname, domain string) string {
	// Normalize: lowercase for case-insensitive matching
	hostname = strings.ToLower(hostname)
	domain = strings.ToLower(domain)

	// v3.6: Fingerprint over FQDN (hostname.domain), not just domain.
	fqdnHash := fqdnFingerprint(hostname, domain)

	// Sanitize hostname: replace invalid chars with hyphens
	sanitizedHostname := sanitizeForDNS(hostname)

	// Combine hostname with hash (Human-readable prefix + unique suffix)
	label := fmt.Sprintf("%s-%s", sanitizedHostname, fqdnHash)

	// DNS-Label max 63 chars (RFC 1123)
	if len(label) > MaxDNSLabelLength {
		// Truncate hostname to fit, keeping the hash suffix intact
		maxHostLen := MaxDNSLabelLength - HashSuffixLength - 1 // -1 for dash
		if maxHostLen < 1 {
			maxHostLen = 1
		}
		truncatedHostname := sanitizedHostname
		if len(sanitizedHostname) > maxHostLen {
			truncatedHostname = sanitizedHostname[:maxHostLen]
		}
		// Remove trailing hyphens after truncation
		truncatedHostname = strings.TrimRight(truncatedHostname, "-")
		label = fmt.Sprintf("%s-%s", truncatedHostname, fqdnHash)
		log.Debugf("DNSLabel truncated: %s (from hostname %s)", label, hostname)
	}

	return label
}

func fqdnFingerprint(hostname, domain string) string {
	mac := hmac.New(sha256.New, []byte(dnsLabelHMACKey))
	_, _ = mac.Write([]byte(hostname + "." + domain))
	sum := mac.Sum(nil)
	return hex.EncodeToString(sum[:hashSuffixBytes])
}

// ValidateDNSLabel checks if a label is RFC 1123 compliant.
// Returns nil if valid, error otherwise.
func ValidateDNSLabel(label string) error {
	if len(label) == 0 {
		return fmt.Errorf("DNS label cannot be empty")
	}
	if len(label) > MaxDNSLabelLength {
		return fmt.Errorf("DNS label must be 1-%d chars, got %d", MaxDNSLabelLength, len(label))
	}

	// RFC 1123: [a-z0-9]([-a-z0-9]*[a-z0-9])?
	// Must be lowercase, start/end with alphanumeric, can contain hyphens
	if !dnsLabelRegex.MatchString(label) {
		return fmt.Errorf("DNS label must match RFC 1123: start/end with alphanumeric, only lowercase letters, digits and hyphens allowed")
	}

	return nil
}

// sanitizeForDNS converts a hostname to a valid DNS label component.
// Replaces invalid characters with hyphens and ensures valid format.
func sanitizeForDNS(hostname string) string {
	// Replace underscores and other common invalid chars with hyphens
	result := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-':
			return r
		case r >= 'A' && r <= 'Z':
			return r + 32 // lowercase
		case r == '_' || r == '.' || r == ' ':
			return '-'
		default:
			return -1 // drop other chars
		}
	}, hostname)

	// Remove leading/trailing hyphens
	result = strings.Trim(result, "-")

	// Collapse multiple consecutive hyphens
	for strings.Contains(result, "--") {
		result = strings.ReplaceAll(result, "--", "-")
	}

	// If empty after sanitization, use a default
	if result == "" {
		result = "peer"
	}

	return result
}

// CheckDNSLabelCollision is a helper that logs a warning if a collision is detected.
// This should be called after DB check for existing label.
// Returns true if collision detected (existingLabel is not empty).
func CheckDNSLabelCollision(label, existingPeerID string) bool {
	if existingPeerID != "" {
		log.Warnf("RARE: DNSLabel collision detected for label %s (existing peer: %s)", label, existingPeerID)
		return true
	}
	return false
}
