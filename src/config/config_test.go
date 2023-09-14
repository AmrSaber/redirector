package config

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

	err := configs.Validate()
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
		t.Errorf("expected error on 'from', got nil")
	}

	// Test invalid "to"
	configs = Config{
		Redirects: []Redirect{
			{From: "example.com", To: "target.com", PreservePath: false},
		},
	}

	err = configs.Validate()
	if err == nil {
		t.Errorf("expected error on 'to', got nil")
	}

	// Test invalid "preserve-path"
	configs = Config{
		Redirects: []Redirect{
			{From: "example.com", To: "https://target.com/some-path", PreservePath: true},
		},
	}

	err = configs.Validate()
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

	err = configs.Validate()
	if err != nil {
		t.Errorf(`unexpected error on "to" wildcard path: %q`, err)
	}

	// Different count
	configs = Config{
		Redirects: []Redirect{
			{From: "example.com", To: "https://*.target.com"},
		},
	}

	err = configs.Validate()
	if err == nil {
		t.Errorf(`expected error on "to" wildcard path, got nil`)
	}
}

func TestConfigDomainMatching(t *testing.T) {
	type Domain string
	mapper := func(d Domain) string { return string(d) }

	// Test exact match
	list := []Domain{"example.com"}

	if matchDomain("example.com", list, mapper) == nil {
		t.Errorf("expected redirect, got nil")
	}

	if matchDomain("example-2.com", list, mapper) != nil {
		t.Errorf("expected nil, got redirect")
	}

	// Test wildcard match
	list = []Domain{"*.example.com"}

	if matchDomain("a.example.com", list, mapper) == nil {
		t.Errorf("expected redirect, got nil")
	}

	if matchDomain("b.example.com", list, mapper) == nil {
		t.Errorf("expected redirect, got nil")
	}

	if matchDomain("a.b.example.com", list, mapper) != nil {
		t.Errorf("expected nil, got redirect")
	}

	if matchDomain("example.com", list, mapper) != nil {
		t.Errorf("expected nil, got redirect")
	}

	// Test it matches first redirect
	list = []Domain{"exact.example.com", "*.example.com"}

	exact := matchDomain("exact.example.com", list, mapper)
	if exact == nil {
		t.Errorf("expected redirect, got nil")
	}

	if exact != &list[0] {
		t.Errorf("expected %s, got %s", list[0], *exact)
	}

	if matchDomain("b.example.com", list, mapper) == nil {
		t.Errorf("expected redirect, got nil")
	}

	if matchDomain("a.b.example.com", list, mapper) != nil {
		t.Errorf("expected nil, got redirect")
	}

	if matchDomain("example.com", list, mapper) != nil {
		t.Errorf("expected nil, got redirect")
	}

	// Test matching exact match first
	list = []Domain{"*.example.com", "exact.example.com"}

	exact = matchDomain("exact.example.com", list, mapper)
	if exact == nil {
		t.Errorf("expected redirect, got nil")
	}

	if exact != &list[1] {
		t.Errorf("expected %s, got %s", list[1], *exact)
	}
}
