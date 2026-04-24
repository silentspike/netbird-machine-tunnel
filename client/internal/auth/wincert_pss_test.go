package auth

import (
	"crypto"
	"crypto/rsa"
	"strings"
	"testing"
)

func TestNcryptPSSSaltLength(t *testing.T) {
	tests := []struct {
		name       string
		opts       *rsa.PSSOptions
		want       uint32
		wantErr    bool
		errMessage string
	}{
		{
			name: "equals hash uses digest size",
			opts: &rsa.PSSOptions{
				Hash:       crypto.SHA256,
				SaltLength: rsa.PSSSaltLengthEqualsHash,
			},
			want: uint32(crypto.SHA256.Size()),
		},
		{
			name: "auto uses digest size for NCrypt TLS signing",
			opts: &rsa.PSSOptions{
				Hash:       crypto.SHA384,
				SaltLength: rsa.PSSSaltLengthAuto,
			},
			want: uint32(crypto.SHA384.Size()),
		},
		{
			name: "explicit salt length",
			opts: &rsa.PSSOptions{
				Hash:       crypto.SHA512,
				SaltLength: 42,
			},
			want: 42,
		},
		{
			name: "unsupported negative salt length",
			opts: &rsa.PSSOptions{
				Hash:       crypto.SHA256,
				SaltLength: -2,
			},
			wantErr:    true,
			errMessage: "unsupported RSA-PSS salt length",
		},
		{
			name: "unsupported hash",
			opts: &rsa.PSSOptions{
				Hash:       crypto.Hash(0),
				SaltLength: rsa.PSSSaltLengthEqualsHash,
			},
			wantErr:    true,
			errMessage: "unsupported hash function",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ncryptPSSSaltLength(tt.opts)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if tt.errMessage != "" && !strings.Contains(err.Error(), tt.errMessage) {
					t.Fatalf("expected error containing %q, got %q", tt.errMessage, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("salt length = %d, want %d", got, tt.want)
			}
		})
	}
}
