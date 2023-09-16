package config

import "testing"

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
