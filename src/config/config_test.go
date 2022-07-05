package config

import (
	"testing"
)

func TestConfigFromYaml(t *testing.T) {
	yml :=
		`
redirects:
  - from: example.com
    to: https://target.com
    preserve-path: false

  - from: "*.example.com"
    to: https://general-target.com
    preserve-path: true
`

	config, err := ConfigFromYaml([]byte(yml))
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	if len(config.Redirects) != 2 {
		t.Errorf("expected 2 redirects, got %d", len(config.Redirects))
	}

	wanted := Config{
		Redirects: []Redirect{
			{From: "example.com", To: "https://target.com", PreservePath: false},
			{From: "*.example.com", To: "https://general-target.com", PreservePath: true},
		},
	}

	for i, r := range wanted.Redirects {
		got := config.Redirects[i]
		if r.From != got.From {
			t.Errorf("expected 'from' %s, got %s", r.From, got.From)
		}

		if r.To != got.To {
			t.Errorf("expected 'to' %s, got %s", r.To, got.To)
		}

		if r.PreservePath != got.PreservePath {
			t.Errorf("expected 'preserve-path' %t, got %t", r.PreservePath, got.PreservePath)
		}
	}
}

func TestConfigValidation(t *testing.T) {
	// Test happy scenario
	configs := Config{
		Redirects: []Redirect{
			{From: "example.com", To: "https://target.com", PreservePath: false},
		},
	}

	err := configs.Validate()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// Allow domain with port
	configs = Config{
		Redirects: []Redirect{
			{From: "example.com:8080", To: "https://target.com", PreservePath: false},
		},
	}

	err = configs.Validate()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// Test invalid "from"
	configs = Config{
		Redirects: []Redirect{
			{From: "example", To: "https://target.com", PreservePath: false},
		},
	}

	err = configs.Validate()
	if err == nil {
		t.Errorf("unexpected error on 'from', got nil")
	}

	// Test invalid "to"
	configs = Config{
		Redirects: []Redirect{
			{From: "example.com", To: "target.com", PreservePath: false},
		},
	}

	err = configs.Validate()
	if err == nil {
		t.Errorf("unexpected error on 'to', got nil")
	}

	// Test invalid "preserve-path"
	configs = Config{
		Redirects: []Redirect{
			{From: "example.com", To: "https://target.com/some-path", PreservePath: true},
		},
	}

	err = configs.Validate()
	if err == nil {
		t.Errorf("unexpected error on 'preserve-path', got nil")
	}
}

func TestConfigDomainMatching(t *testing.T) {
	// Test exact match
	configs := Config{
		Redirects: []Redirect{
			{From: "example.com", To: "https://target.com", PreservePath: false},
		},
	}

	if configs.GetRedirect("example.com") == nil {
		t.Errorf("expected redirect, got nil")
	}

	if configs.GetRedirect("example-2.com") != nil {
		t.Errorf("expected nil, got redirect")
	}

	// Test wildcard match
	configs = Config{
		Redirects: []Redirect{
			{From: "*.example.com", To: "https://target.com", PreservePath: true},
		},
	}

	if configs.GetRedirect("a.example.com") == nil {
		t.Errorf("expected redirect, got nil")
	}

	if configs.GetRedirect("b.example.com") == nil {
		t.Errorf("expected redirect, got nil")
	}

	if configs.GetRedirect("a.b.example.com") != nil {
		t.Errorf("expected nil, got redirect")
	}

	if configs.GetRedirect("example.com") != nil {
		t.Errorf("expected nil, got redirect")
	}

	// Test it matches first redirect
	configs = Config{
		Redirects: []Redirect{
			{From: "exact.example.com", To: "https://exact-target.com", PreservePath: true},
			{From: "*.example.com", To: "https://target.com", PreservePath: true},
		},
	}

	exact := configs.GetRedirect("exact.example.com")
	if exact == nil {
		t.Errorf("expected redirect, got nil")
	}

	if exact != nil && exact.To != configs.Redirects[0].To {
		t.Errorf("expected %s, got %s", configs.Redirects[0].To, exact.To)
	}

	if configs.GetRedirect("b.example.com") == nil {
		t.Errorf("expected redirect, got nil")
	}

	if configs.GetRedirect("a.b.example.com") != nil {
		t.Errorf("expected nil, got redirect")
	}

	if configs.GetRedirect("example.com") != nil {
		t.Errorf("expected nil, got redirect")
	}

	// Test matching wildcard first
	configs = Config{
		Redirects: []Redirect{
			{From: "*.example.com", To: "https://target.com", PreservePath: true},
			{From: "exact.example.com", To: "https://exact-target.com", PreservePath: true},
		},
	}

	exact = configs.GetRedirect("exact.example.com")
	if exact == nil {
		t.Errorf("expected redirect, got nil")
	}

	if exact != nil && exact.To != configs.Redirects[0].To {
		t.Errorf("expected %s, got %s", configs.Redirects[0].To, exact.To)
	}
}
