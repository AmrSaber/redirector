package models

import (
	"testing"
)

func TestConfigValidation(t *testing.T) {
	// Test happy scenario
	configs := Config{
		Redirects: []Redirect{
			{From: "example.com", To: "https://target.com", PreservePath: false},
		},
	}

	err := configs.validate()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	// Test invalid "from"
	configs = Config{
		Redirects: []Redirect{
			{From: "example", To: "https://target.com", PreservePath: false},
		},
	}

	err = configs.validate()
	if err == nil {
		t.Errorf("expected error on 'from', got nil")
	}

	// Test invalid "to"
	configs = Config{
		Redirects: []Redirect{
			{From: "example.com", To: "target.com", PreservePath: false},
		},
	}

	err = configs.validate()
	if err == nil {
		t.Errorf("expected error on 'to', got nil")
	}

	// Test invalid "preserve-path"
	configs = Config{
		Redirects: []Redirect{
			{From: "example.com", To: "https://target.com/some-path", PreservePath: true},
		},
	}

	err = configs.validate()
	if err == nil {
		t.Errorf("expected error on 'preserve-path', got nil")
	}

	// Test "to" wildcards
	// same count
	configs = Config{
		Redirects: []Redirect{
			{From: "*.*.example.com", To: "https://a.*.target.com"},
		},
	}

	err = configs.validate()
	if err != nil {
		t.Errorf(`unexpected error on "to" wildcard path: %q`, err)
	}

	// Different count
	configs = Config{
		Redirects: []Redirect{
			{From: "example.com", To: "https://*.target.com"},
		},
	}

	err = configs.validate()
	if err == nil {
		t.Errorf(`expected error on "to" wildcard path, got nil`)
	}
}
