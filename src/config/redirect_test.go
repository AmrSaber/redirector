package config

import (
	"net/http"
	"testing"
)

func TestResolveDomain(t *testing.T) {
	// Test wildcard redirection
	redirect := Redirect{
		From: "*.amr-saber.com",
		To:   "https://*.amrsaber.io",
	}

	request, _ := http.NewRequest("GET", "https://abc.amr-saber.com", nil)
	got := redirect.ResolvePath(request)
	expected := "https://abc.amrsaber.io"
	if got != expected {
		t.Errorf("got %q, expected %q", got, expected)
	}

	// Test preserve path
	redirect = Redirect{
		From:         "amr-saber.com",
		To:           "https://amrsaber.io",
		PreservePath: true,
	}

	request, _ = http.NewRequest("GET", "https://amr-saber.com/some-path", nil)
	got = redirect.ResolvePath(request)
	expected = "https://amrsaber.io/some-path"
	if got != expected {
		t.Errorf("got %q, expected %q", got, expected)
	}

	// Test preserve path with wildcard redirection
	redirect = Redirect{
		From:         "*.amr-saber.com",
		To:           "https://*.amrsaber.io",
		PreservePath: true,
	}

	request, _ = http.NewRequest("GET", "https://abc.amr-saber.com/some-path", nil)
	got = redirect.ResolvePath(request)
	expected = "https://abc.amrsaber.io/some-path"
	if got != expected {
		t.Errorf("got %q, expected %q", got, expected)
	}
}
