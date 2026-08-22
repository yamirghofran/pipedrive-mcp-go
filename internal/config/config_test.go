package config

import "testing"

func TestNormalizeDomain(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    string
		wantErr bool
	}{
		{name: "plain domain", raw: "acme.pipedrive.com", want: "acme.pipedrive.com"},
		{name: "https scheme prefix", raw: "https://acme.pipedrive.com", want: "acme.pipedrive.com"},
		{name: "http scheme prefix", raw: "http://acme.pipedrive.com", want: "acme.pipedrive.com"},
		{name: "trailing slash", raw: "acme.pipedrive.com/", want: "acme.pipedrive.com"},
		{name: "full API URL", raw: "https://acme.pipedrive.com/api/v1/deals", want: "acme.pipedrive.com"},
		{name: "mixed case", raw: "HTTPS://Acme.Pipedrive.com", want: "acme.pipedrive.com"},
		{name: "bare company subdomain", raw: "acme", want: "acme.pipedrive.com"},
		{name: "surrounding whitespace", raw: "  acme.pipedrive.com  ", want: "acme.pipedrive.com"},
		{name: "empty", raw: "", wantErr: true},
		{name: "whitespace only", raw: "   ", wantErr: true},
		{name: "scheme only", raw: "https://", wantErr: true},
		{name: "scheme and slashes only", raw: "https:///", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeDomain(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("normalizeDomain(%q) = %q, want error", tt.raw, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("normalizeDomain(%q) returned error: %v", tt.raw, err)
			}
			if got != tt.want {
				t.Fatalf("normalizeDomain(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestLoadNormalizesDomain(t *testing.T) {
	t.Setenv("PIPEDRIVE_API_TOKEN", "test-token")
	t.Setenv("PIPEDRIVE_DOMAIN", "https://acme.pipedrive.com/")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	if cfg.PipedriveDomain != "acme.pipedrive.com" {
		t.Fatalf("cfg.PipedriveDomain = %q, want %q", cfg.PipedriveDomain, "acme.pipedrive.com")
	}
}

func TestLoadRejectsInvalidDomain(t *testing.T) {
	t.Setenv("PIPEDRIVE_API_TOKEN", "test-token")
	t.Setenv("PIPEDRIVE_DOMAIN", "https://")

	if _, err := Load(); err == nil {
		t.Fatal(`Load() succeeded with domain "https://", want error`)
	}
}

func TestLoadRequiresToken(t *testing.T) {
	t.Setenv("PIPEDRIVE_API_TOKEN", "")
	t.Setenv("PIPEDRIVE_DOMAIN", "acme.pipedrive.com")

	if _, err := Load(); err == nil {
		t.Fatal("Load() succeeded without PIPEDRIVE_API_TOKEN, want error")
	}
}
