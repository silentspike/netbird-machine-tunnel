package mtls

import (
	"strings"
	"testing"
)

func TestValidateIssuerCA(t *testing.T) {
	const (
		accountID   = "account-123"
		issuerFP    = "abcdef0123456789"
		otherIssuer = "1234567890abcdef"
	)

	t.Run("matching issuer is accepted", func(t *testing.T) {
		SetValidatorConfig(&ValidatorConfig{
			AccountAllowedIssuers: map[string][]string{
				accountID: {issuerFP},
			},
		})
		defer SetValidatorConfig(nil)

		if err := ValidateIssuerCA(accountID, strings.ToUpper(issuerFP)); err != nil {
			t.Fatalf("ValidateIssuerCA returned error: %v", err)
		}
	})

	t.Run("wrong issuer is rejected", func(t *testing.T) {
		SetValidatorConfig(&ValidatorConfig{
			AccountAllowedIssuers: map[string][]string{
				accountID: {issuerFP},
			},
		})
		defer SetValidatorConfig(nil)

		err := ValidateIssuerCA(accountID, otherIssuer)
		if err == nil {
			t.Fatal("expected wrong issuer to be rejected")
		}
		if !strings.Contains(err.Error(), "not in allowed list") {
			t.Fatalf("error = %q, want allowlist rejection", err.Error())
		}
	})

	t.Run("empty account allowlist is fail closed", func(t *testing.T) {
		SetValidatorConfig(&ValidatorConfig{
			AccountAllowedIssuers: map[string][]string{
				accountID: {},
			},
		})
		defer SetValidatorConfig(nil)

		err := ValidateIssuerCA(accountID, issuerFP)
		if err == nil {
			t.Fatal("expected empty issuer allowlist to be rejected")
		}
		if !strings.Contains(err.Error(), "no allowed CA issuers configured") {
			t.Fatalf("error = %q, want empty allowlist rejection", err.Error())
		}
	})
}
